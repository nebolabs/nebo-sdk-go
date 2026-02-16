# Nebo Go SDK

Official Go SDK for building [Nebo](https://neboloop.com) apps.

## Install

```bash
go get github.com/nebolabs/nebo-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    nebo "github.com/nebolabs/nebo-sdk-go"
)

type Calculator struct{}

func (c *Calculator) Name() string        { return "calculator" }
func (c *Calculator) Description() string { return "Performs arithmetic calculations." }

func (c *Calculator) Schema() json.RawMessage {
    return nebo.NewSchema("add", "subtract", "multiply", "divide").
        Number("a", "First operand", true).
        Number("b", "Second operand", true).
        Build()
}

func (c *Calculator) Execute(_ context.Context, input json.RawMessage) (string, error) {
    var in struct {
        Action string  `json:"action"`
        A      float64 `json:"a"`
        B      float64 `json:"b"`
    }
    if err := json.Unmarshal(input, &in); err != nil {
        return "", fmt.Errorf("invalid input: %w", err)
    }

    switch in.Action {
    case "add":
        return fmt.Sprintf("%g", in.A+in.B), nil
    case "subtract":
        return fmt.Sprintf("%g", in.A-in.B), nil
    case "multiply":
        return fmt.Sprintf("%g", in.A*in.B), nil
    case "divide":
        if in.B == 0 {
            return "", fmt.Errorf("division by zero")
        }
        return fmt.Sprintf("%g", in.A/in.B), nil
    default:
        return "", fmt.Errorf("unknown action: %s", in.Action)
    }
}

func main() {
    app, err := nebo.New()
    if err != nil {
        log.Fatal(err)
    }
    app.RegisterTool(&Calculator{})
    log.Fatal(app.Run())
}
```

## Handler Interfaces

| Capability | Interface | Manifest `provides` |
|-----------|-----------|---------------------|
| Tool | `ToolHandler` | `tool:<name>` |
| Channel | `ChannelHandler` | `channel:<name>` |
| Gateway | `GatewayHandler` | `gateway` |
| UI | `UIHandler` | `ui` |
| Comm | `CommHandler` | `comm` |
| Schedule | `ScheduleHandler` | `schedule` |

## Schema Builder

Build JSON Schema for STRAP-pattern tool inputs:

```go
schema := nebo.NewSchema("list", "create", "delete").
    String("name", "Resource name", true).
    Number("limit", "Max results", false).
    Bool("verbose", "Show details", false).
    Enum("format", "Output format", false, "json", "text").
    Build()
```

## View Builder

Build structured UI views:

```go
view := nebo.NewView("dashboard", "My Dashboard").
    Heading("title", "Status", "h2").
    Text("info", "All systems operational").
    Button("refresh", "Refresh", "primary").
    Divider("sep").
    Input("search", "", "Search...").
    Build()
```

## Documentation

See [Creating Nebo Apps](https://neboloop.com/developers) for the full guide.
