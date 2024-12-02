```mermaid
sequenceDiagram
    participant DB as Database 
    participant A as Agent 
    participant Q as Agent Internal Queue
    participant C as Collector 
    participant W as DataWarehouse 
    
    A-->>DB: Query active sessions
    DB-->>A: Provide session/query data
    A-->>A:Proccess query samples
    A-->>A:CreateSnapshot
    A-->>Q:Send normalized queries to be monitored
    A-->>C: Send Snapshot
    C-->>W: Save Snapshot
    Q-->>A: send normalized queries
    A-->>DB:Query stats on normalized queries
    DB-->>A: send NQ stats
    A-->>A: preprocess NQ stats
    A-->>C:send NQ stats
    C-->>W:store NQ stats
    
    
```

## databases
- active connections
- blocking history
- session graphs
- top queries
- deadlocks

## db snapshots
- samples from snapshot
- blocking tree
- wait events
- time elapsed

## sample
- link to normalized query
- details
- exec plan
- blocking tree

## normalized query
- statistics
- link to samples
- execution plan history
- blocking history

