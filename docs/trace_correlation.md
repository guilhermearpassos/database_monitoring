# Trace ID Correlation (SQL Server `CONTEXT_INFO` + Grafana Tempo)

Sqlsights can correlate live SQL Server query snapshots, blocking chains, and problematic execution plans directly back to the distributed application trace that triggered them—letting you jump straight from a slow or blocking SQL query into **Grafana Tempo**.

---

## How It Works

1. Your application packs its active OpenTelemetry W3C Trace Context (`TraceID`, `SpanID`, `Flags`) into a compact binary payload prefixed with the 4-byte ASCII magic header `TPRC` and sets it on the SQL Server session via `SET CONTEXT_INFO`.
2. When the Sqlsights agent polls SQL Server DMVs (`sys.dm_exec_requests.context_info` and `sys.dm_exec_sessions.context_info`), it detects the `TPRC` header and decodes the 32-character hex `trace_id`.
3. The ingester stores `trace_id` indexed in PostgreSQL (`query_samples.trace_id`).
4. In the Grafana plugin (**Snapshots**, **Plan Analysis**, **Blocking Chains**, and **Query Details**), any sample carrying a `trace_id` displays a clickable badge that opens **Grafana Explore → Tempo** directly to that trace.

---

## Binary Payload Format (`TPRC` v0)

SQL Server's `CONTEXT_INFO` holds up to **128 bytes** (`varbinary(128)`). Sqlsights expects at least **30 bytes** (zero-padding up to 128 bytes is optional):

| Offset | Length | Field | Value / Encoding |
| :--- | :--- | :--- | :--- |
| `0..3` | 4 bytes | Magic Header | ASCII `TPRC` (`0x54 0x50 0x52 0x43`) |
| `4` | 1 byte | Version | `0x00` (W3C Trace Context v0) |
| `5..20` | 16 bytes | Trace ID | Raw 16-byte OpenTelemetry / W3C Trace ID |
| `21..28` | 8 bytes | Span ID | Raw 8-byte Span ID |
| `29` | 1 byte | Trace Flags | 1-byte W3C trace flags (e.g., `0x01` = sampled) |
| `30..127` | 0–98 bytes | Padding | Optional `0x00` padding |

---

## Instrumenting Your Application

### Go (`database/sql` / `sqlx` + `go-mssqldb` + OpenTelemetry)

```go
package dbtrace

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

var magicTraceparent = []byte("TPRC")

// PackTraceContext encodes the active OpenTelemetry span context into a
// SQL Server CONTEXT_INFO payload recognized by Sqlsights.
func PackTraceContext(ctx context.Context) []byte {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return nil
	}

	buf := make([]byte, 30)
	copy(buf[0:4], magicTraceparent) // Magic header "TPRC"
	buf[4] = 0x00                    // Version 0
	tid := sc.TraceID()
	copy(buf[5:21], tid[:])          // 16-byte TraceID
	sid := sc.SpanID()
	copy(buf[21:29], sid[:])         // 8-byte SpanID
	buf[29] = byte(sc.TraceFlags())  // 1-byte Flags
	return buf
}
```

Pass the packed bytes in the same batch (or via a `driver.Conn` wrapper) right before your statement:

```go
ctxInfo := dbtrace.PackTraceContext(ctx)
if len(ctxInfo) > 0 {
	_, err := db.ExecContext(ctx, "SET CONTEXT_INFO @p1; SELECT ... FROM orders WHERE id = @p2", ctxInfo, orderID)
}
```

### .NET / C# (`Microsoft.Data.SqlClient` + `System.Diagnostics.Activity`)

```csharp
using System;
using System.Diagnostics;

public static class SqlTraceContext
{
    private static readonly byte[] Magic = "TPRC"u8.ToArray();

    public static byte[]? PackCurrentTraceContext()
    {
        var activity = Activity.Current;
        if (activity is null || activity.TraceId == default)
            return null;

        var buf = new byte[30];
        Magic.CopyTo(buf, 0);                            // 0..3: "TPRC"
        buf[4] = 0x00;                                   // 4: Version 0
        activity.TraceId.CopyTo(buf.AsSpan(5, 16));      // 5..20: TraceID (16 bytes)
        activity.SpanId.CopyTo(buf.AsSpan(21, 8));       // 21..28: SpanID (8 bytes)
        buf[29] = (byte)activity.ActivityTraceFlags;     // 29: Flags
        return buf;
    }
}
```

You can set it inside an EF Core `DbCommandInterceptor` (`ReaderExecutingAsync` / `NonQueryExecutingAsync`) or before Dapper queries by prepending `SET CONTEXT_INFO @__sqlsights_ctx;` and adding a `SqlDbType.VarBinary` parameter of size `128`.

### Node.js / TypeScript (`@opentelemetry/api` + `mssql`)

```ts
import { context, trace } from '@opentelemetry/api';

export function packTraceContext(): Buffer | null {
  const spanContext = trace.getSpan(context.active())?.spanContext();
  if (!spanContext) {
    return null;
  }

  const buf = Buffer.alloc(30);
  buf.write('TPRC', 0, 'ascii');                       // 0..3: Magic header
  buf.writeUInt8(0, 4);                                // 4: Version 0
  Buffer.from(spanContext.traceId, 'hex').copy(buf, 5); // 5..20: TraceID (16 bytes)
  Buffer.from(spanContext.spanId, 'hex').copy(buf, 21); // 21..28: SpanID (8 bytes)
  buf.writeUInt8(spanContext.traceFlags, 29);          // 29: Flags
  return buf;
}
```

> **Connection Pooling Note**
> `CONTEXT_INFO` is scoped to the SQL Server **session**. If your application reuses pooled connections:
> - Set `CONTEXT_INFO` at the start of every request/batch (so it always reflects the current trace), **and**
> - Clear it with `SET CONTEXT_INFO 0x;` when a connection is checked back into the pool (or when executing untraced background queries) so an older `trace_id` isn't attributed to later queries on the same session.

---

## Configuring Grafana & Tempo

### Local Development (`docker compose`)

The local `docker-compose.yml` stack ships pre-wired with **Grafana Tempo**:

- **`tempo`** (`grafana/tempo:2.7.2`, config in `tempo.yaml`) listens on HTTP `:3200` (for Grafana queries) and OTLP gRPC `:4317` (internal).
- **`otel-collector`** (`otel-collector-config.yaml`) fans out incoming OTLP traces on `localhost:4317` to **Tempo**, Jaeger, and Zipkin simultaneously.
- **`provisioning/datasources/datasources.yml`** auto-provisions the built-in Grafana Tempo datasource (`uid: tempo`, `url: http://tempo:3200`).
- **`provisioning/plugins/apps.yaml`** pre-configures `tempoDatasourceUid: tempo` on the Sqlsights app plugin.

Point your application's OTLP exporter at `localhost:4317` (or `otel-collector:4317` inside Docker), start the stack (`docker compose up -d`), and traces will appear in Tempo automatically.

### Production Grafana Setup

1. Add a **Tempo** datasource in Grafana (**Connections → Data sources → Add data source → Tempo**).
2. Open **Apps → Sqlsights → Configuration**.
3. Under **Tracing Integration (Tempo)**, pick your Tempo datasource from the dropdown and click **Save API settings**.
   - If left blank, Sqlsights automatically falls back to the first configured `tempo` datasource in Grafana.

---

## Quick End-to-End Verification

You can test `CONTEXT_INFO` extraction directly from SSMS / `sqlcmd` against a monitored SQL Server instance:

```sql
-- Magic "TPRC" (54505243) + Version (00)
-- + TraceID (0123456789abcdef0123456789abcdef)
-- + SpanID  (0123456789abcdef)
-- + Flags   (01)
SET CONTEXT_INFO 0x54505243000123456789ABCDEF0123456789ABCDEF0123456789ABCDEF01;
WAITFOR DELAY '00:00:15';
SET CONTEXT_INFO 0x;
```

Within one snapshot interval, open **Snapshots** in Grafana, expand the captured snapshot, and you will see the clickable `01234567…` Tempo trace badge next to the `WAITFOR DELAY` sample.
