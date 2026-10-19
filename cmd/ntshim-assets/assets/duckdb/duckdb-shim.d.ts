// ntshim/duckdb — TypeScript types for the JS bridge.

declare global {
    interface Window {
        ntshim?: NTShim;
    }

    /** Declaration-merged across every ntshim bridge type file —
     *  each shim contributes its own optional member. */
    interface NTShim {
        duckdb?: NTShimDuckDB;
    }
}

export interface NTShimDuckDB {
    /** Resolves once the DuckDB-Wasm worker is up and ready for open(). */
    ready: Promise<void>;

    /** Opens a DuckDB connection from a DSN ("path?key=val" or ":memory:").
     *  Returns an opaque handle the Go driver passes back on every call. */
    open(dsn: string): Promise<number>;

    /** Closes a connection. */
    close(handle: number): Promise<void>;

    /** Executes SQL and returns Arrow IPC stream bytes for SELECT results. */
    query(handle: number, sql: string, params: unknown[]): Promise<Uint8Array>;

    /** Executes SQL for DDL/DML; rowsAffected is always 0 (DuckDB-Wasm
     *  does not expose it). */
    exec(handle: number, sql: string, params: unknown[]): Promise<{ rowsAffected: number }>;

    /** Registers a remote URL as a virtual file for ATTACH / read_parquet / etc. */
    registerFileURL(name: string, url: string): Promise<void>;

    /** Registers an in-memory buffer as a virtual file. */
    registerFileBuffer(name: string, bytes: Uint8Array): Promise<void>;
}

export {};
