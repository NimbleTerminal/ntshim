// ntshim/pdfium — TypeScript types for the JS bridge.

declare global {
    interface Window {
        ntshim?: NTShim;
    }

    /** Declaration-merged across every ntshim bridge type file —
     *  each shim contributes its own optional member. */
    interface NTShim {
        pdfium?: NTShimPDFium;
    }
}

export interface NTShimPDFium {
    /** Resolves once the PDFium WebAssembly module is initialized. */
    ready: Promise<void>;

    /** Loads a PDF document from an in-memory buffer. Returns an opaque handle. */
    loadDocument(uint8: Uint8Array): Promise<number>;

    /** Closes a document handle and frees its memory. */
    closeDocument(handle: number): void;

    /** Returns the total page count for the loaded document. */
    numPages(handle: number): number;

    /** Returns the page dimensions in points (1/72 inch). */
    pageSize(handle: number, pageIdx: number): { widthPt: number, heightPt: number };

    /** Rasterizes a page at the specified DPI. Returns RGBA buffer and dimensions. */
    renderPage(handle: number, pageIdx: number, dpi: number): Promise<{
        data: Uint8ClampedArray;
        width: number;
        height: number;
    }>;
}

export {};
