package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"dataSerializationForCache/internal/invoice"
	"dataSerializationForCache/internal/serializer"
	"dataSerializationForCache/pkg/kvstore"
)

type openFunc func() (kvstore.Store, error)

// storeCmd serializes the sample invoice in one format and writes it to Valkey.
func storeCmd(open openFunc) *cobra.Command {
	var (
		format, id string
		ttl        time.Duration
	)
	cmd := &cobra.Command{
		Use:     "store",
		Short:   "Serialize the sample invoice and store it in Valkey",
		Example: "  cli store --format avro --id INV-1001 --ttl 10m",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			codec, err := serializer.ByName(format)
			if err != nil {
				return err
			}
			data, err := codec.Marshal(invoice.Sample(id))
			if err != nil {
				return fmt.Errorf("marshal %s: %w", codec.Name(), err)
			}
			store, err := open()
			if err != nil {
				return err
			}
			defer store.CloseWithCheck(cmd.Context())

			key := cacheKey(codec.Name(), id)
			// kvstore.Set formats values with fmt.Sprint, which would turn a []byte into "[1 2 3]".
			// A string keeps the bytes unchanged.
			if err = store.Set(cmd.Context(), key, string(data), ttl); err != nil {
				return fmt.Errorf("store %s: %w", key, err)
			}
			ttlText := "none"
			if ttl > 0 {
				ttlText = ttl.String()
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "stored %s (%d bytes, ttl %s)\n", key, len(data), ttlText)
			return err
		},
	}
	addFormatFlag(cmd, &format)
	cmd.Flags().StringVar(&id, "id", "INV-1001", "invoice ID")
	cmd.Flags().DurationVar(&ttl, "ttl", 10*time.Minute, "cache entry TTL (0 = no expiration)")
	return cmd
}

// fetchCmd reads an invoice from Valkey, decodes it and reports how long GET and decode took.
// A single measurement is noisy; compare averages over many runs.
func fetchCmd(open openFunc) *cobra.Command {
	var format, id string
	cmd := &cobra.Command{
		Use:     "fetch",
		Short:   "Fetch an invoice from Valkey and deserialize it",
		Example: "  cli fetch --format avro --id INV-1001",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			codec, err := serializer.ByName(format)
			if err != nil {
				return err
			}
			store, err := open()
			if err != nil {
				return err
			}
			defer store.CloseWithCheck(cmd.Context())

			key := cacheKey(codec.Name(), id)
			start := time.Now()
			cached, err := store.Get(cmd.Context(), key)
			getTime := time.Since(start)
			if errors.Is(err, kvstore.ErrNotFound) {
				return fmt.Errorf("%s not found, run `cli store --format %s --id %s` first", key, codec.Name(), id)
			}
			if err != nil {
				return fmt.Errorf("fetch %s: %w", key, err)
			}

			start = time.Now()
			inv, err := codec.Unmarshal([]byte(cached))
			decodeTime := time.Since(start)
			if err != nil {
				return fmt.Errorf("unmarshal %s: %w", codec.Name(), err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(),
				"fetched %s (%d bytes) in %s (get %s, decode %s)\n"+
					"invoice %s for %s issued %s: %d lines, total %.2f %s\n",
				key, len(cached), formatDuration(getTime+decodeTime), formatDuration(getTime), formatDuration(decodeTime),
				inv.ID, inv.CustomerID, inv.IssuedAt.Format("2006-01-02"),
				len(inv.Lines), float64(inv.TotalCents())/100, inv.Currency,
			)
			return err
		},
	}
	addFormatFlag(cmd, &format)
	cmd.Flags().StringVar(&id, "id", "INV-1001", "invoice ID")
	return cmd
}

// compareCmd stores and reads back the invoice in every format --runs times
// and prints the payload size with average timings.
func compareCmd(open openFunc) *cobra.Command {
	var (
		id   string
		ttl  time.Duration
		runs int
	)
	cmd := &cobra.Command{
		Use:     "compare",
		Short:   "Store the invoice in every format and compare payload size and speed",
		Example: "  cli compare --id INV-1001 --runs 1000",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if runs < 1 {
				return fmt.Errorf("--runs must be at least 1, got %d", runs)
			}
			codecs, err := serializer.All()
			if err != nil {
				return err
			}
			store, err := open()
			if err != nil {
				return err
			}
			defer store.CloseWithCheck(cmd.Context())

			want := invoice.Sample(id)
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "invoice %s with %d lines, averages over %d runs\n", id, len(want.Lines), runs)
			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', tabwriter.AlignRight)
			fmt.Fprintln(w, "FORMAT\tBYTES\tENCODE\tSET\tSET TOTAL\tGET\tDECODE\tGET TOTAL\tTOTAL\tROUND-TRIP\t")
			for _, codec := range codecs {
				m, measureErr := measure(cmd, store, codec, want, ttl, runs)
				if measureErr != nil {
					return fmt.Errorf("%s: %w", codec.Name(), measureErr)
				}
				setTotal, getTotal := m.encode+m.set, m.get+m.decode
				fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
					codec.Name(), m.size,
					formatDuration(m.encode), formatDuration(m.set), formatDuration(setTotal),
					formatDuration(m.get), formatDuration(m.decode), formatDuration(getTotal),
					formatDuration(setTotal+getTotal), m.roundTrip)
			}
			return w.Flush()
		},
	}
	cmd.Flags().StringVar(&id, "id", "INV-1001", "invoice ID")
	cmd.Flags().DurationVar(&ttl, "ttl", 10*time.Minute, "cache entry TTL (0 = no expiration)")
	cmd.Flags().IntVar(&runs, "runs", 500, "how many times each step is repeated to average timings")
	return cmd
}

// measurement holds the results for one format. Durations are averages per run.
type measurement struct {
	size                     int
	encode, set, get, decode time.Duration
	roundTrip                string
}

// measure times encode, SET, GET and decode separately, runs times each,
// so network cost and format cost can be compared.
func measure(
	cmd *cobra.Command, store kvstore.Store, codec serializer.Codec,
	want invoice.Invoice, ttl time.Duration, runs int,
) (measurement, error) {
	ctx := cmd.Context()
	key := cacheKey(codec.Name(), want.ID)
	var (
		m    measurement
		data []byte
		got  invoice.Invoice
		err  error
	)
	repeat := func(step func() error) (time.Duration, error) {
		start := time.Now()
		for range runs {
			if stepErr := step(); stepErr != nil {
				return 0, stepErr
			}
		}
		return time.Since(start) / time.Duration(runs), nil
	}

	if m.encode, err = repeat(func() (marshalErr error) {
		data, marshalErr = codec.Marshal(want)
		return marshalErr
	}); err != nil {
		return m, fmt.Errorf("marshal: %w", err)
	}
	m.size = len(data)

	if m.set, err = repeat(func() error {
		return store.Set(ctx, key, string(data), ttl)
	}); err != nil {
		return m, fmt.Errorf("store: %w", err)
	}

	var cached string
	if m.get, err = repeat(func() (getErr error) {
		cached, getErr = store.Get(ctx, key)
		return getErr
	}); err != nil {
		return m, fmt.Errorf("fetch: %w", err)
	}

	if m.decode, err = repeat(func() (unmarshalErr error) {
		got, unmarshalErr = codec.Unmarshal([]byte(cached))
		return unmarshalErr
	}); err != nil {
		return m, fmt.Errorf("unmarshal: %w", err)
	}

	m.roundTrip = "ok"
	if !reflect.DeepEqual(want, got) {
		m.roundTrip = "mismatch"
	}
	return m, nil
}

func addFormatFlag(cmd *cobra.Command, format *string) {
	cmd.Flags().StringVar(format, "format", serializer.FormatJSON,
		"serialization format: "+strings.Join(serializer.Formats(), ", "))
}

// cacheKey includes the format so every encoding of an invoice can be cached side by side.
func cacheKey(format, invoiceID string) string {
	return "invoice:" + format + ":" + invoiceID
}

// formatDuration rounds to a readable precision: 0.1µs below 1ms, 0.01ms above.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return d.Round(100 * time.Nanosecond).String()
	}
	return d.Round(10 * time.Microsecond).String()
}
