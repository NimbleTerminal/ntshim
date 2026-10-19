// ntshim/pdfium — JS bridge file
//
// Loaded by consumer pages before their Go-WASM main runs. Sets
// window.ntshim.pdfium with the API the Go side expects. Renders
// PDF pages to packed RGBA buffers via @embedpdf/pdfium.
//
// Ref: https://www.embedpdf.com/docs/pdfium

import { init as initPDFium } from 'https://cdn.jsdelivr.net/npm/@embedpdf/pdfium@2.15.1/+esm';

const NAMESPACE = (window.ntshim = window.ntshim || {});

// Render-flags bitfield. FPDF_REVERSE_BYTE_ORDER (0x10) makes pdfium
// emit RGBA byte order directly, sidestepping a JS-side byte swap.
const FPDF_REVERSE_BYTE_ORDER = 0x10;
// FPDF_ANNOT also flattens annotations into the render (= what users expect).
const FPDF_ANNOT = 0x01;

// Internal state. Documents map an opaque small-int handle (what we
// hand to Go) to { doc: pdfium-doc-ptr, dataPtr: malloc'd-buffer }.
// We must keep dataPtr alive for the lifetime of the doc — pdfium's
// FPDF_LoadMemDocument keeps a reference into our memory.
const docs = new Map();
let nextHandle = 1;
let wrapped = null;       // WrappedPdfiumModule from initPDFium()

async function bootstrap() {
    wrapped = await initPDFium({
        locateFile: (path) => {
            if (path.endsWith('.wasm')) {
                return 'https://cdn.jsdelivr.net/npm/@embedpdf/pdfium@2.15.1/dist/pdfium.wasm';
            }
            return path;
        }
    });
    // @embedpdf/pdfium's init() instantiates the wasm and wraps the
    // function exports but does NOT call PDFium's own library-init
    // entry points — without these every FPDF_* call reaches into
    // uninitialized state and the first global-pointer dereference
    // traps with "Out of bounds memory access" (FPDF_LoadPage in our
    // case).
    wrapped.FPDF_InitLibrary();
    // PDFiumExt_Init wires up @embedpdf's extension state (form-field
    // handling, helpers used by the EPDF_* wrappers). Cheap and
    // required for completeness.
    wrapped.PDFiumExt_Init();
}

function getDoc(handle) {
    const entry = docs.get(handle);
    if (!entry) throw new Error(`pdfium: invalid document handle ${handle}`);
    return entry;
}

async function loadDocument(uint8) {
    if (!wrapped) throw new Error('pdfium: bridge not bootstrapped (await ntshim.pdfium.ready first)');
    const { pdfium, FPDF_LoadMemDocument64 } = wrapped;
    const size = uint8.byteLength;
    // Copy bytes into pdfium's wasm heap. The buffer must stay alive
    // until FPDF_CloseDocument, so we stash the pointer on the doc
    // entry and free it in closeDocument.
    const dataPtr = pdfium.wasmExports.malloc(size);
    if (!dataPtr) throw new Error(`pdfium: malloc(${size}) failed in pdfium heap`);
    pdfium.HEAPU8.set(uint8, dataPtr);
    const doc = FPDF_LoadMemDocument64(dataPtr, size, '');
    if (!doc) {
        pdfium.wasmExports.free(dataPtr);
        throw new Error('pdfium: FPDF_LoadMemDocument64 returned NULL (unsupported / encrypted PDF?)');
    }
    const handle = nextHandle++;
    docs.set(handle, { doc, dataPtr });
    return handle;
}

function closeDocument(handle) {
    const entry = docs.get(handle);
    if (!entry) return;
    docs.delete(handle);
    try {
        wrapped.FPDF_CloseDocument(entry.doc);
    } finally {
        wrapped.pdfium.wasmExports.free(entry.dataPtr);
    }
}

function numPages(handle) {
    return wrapped.FPDF_GetPageCount(getDoc(handle).doc);
}

// pageSize returns PostScript points (1/72 inch). Uses the F
// (single-float-struct) variant — cheaper than the double-ptr version
// and we only need 32-bit precision here.
function pageSize(handle, pageIdx) {
    const entry = getDoc(handle);
    const { pdfium, FPDF_GetPageSizeByIndexF } = wrapped;
    // Alloc 8 bytes for FS_SIZEF { float width, float height }.
    const sizePtr = pdfium.wasmExports.malloc(8);
    try {
        if (!FPDF_GetPageSizeByIndexF(entry.doc, pageIdx, sizePtr)) {
            throw new Error(`pdfium: FPDF_GetPageSizeByIndexF failed for page ${pageIdx}`);
        }
        const widthPt  = pdfium.getValue(sizePtr,     'float');
        const heightPt = pdfium.getValue(sizePtr + 4, 'float');
        return { widthPt, heightPt };
    } finally {
        pdfium.wasmExports.free(sizePtr);
    }
}

async function renderPage(handle, pageIdx, dpi) {
    const entry = getDoc(handle);
    const {
        pdfium,
        FPDF_LoadPage,
        FPDF_ClosePage,
        FPDFBitmap_Create,
        FPDFBitmap_Destroy,
        FPDFBitmap_FillRect,
        FPDFBitmap_GetBuffer,
        FPDFBitmap_GetStride,
        FPDF_RenderPageBitmap,
    } = wrapped;

    // Resolve page size to compute output pixel dims at the requested
    // DPI.
    const { widthPt, heightPt } = pageSize(handle, pageIdx);
    const widthPx  = Math.max(1, Math.round(widthPt  * dpi / 72));
    const heightPx = Math.max(1, Math.round(heightPt * dpi / 72));

    // Bitmap: FPDFBitmap_Create with alpha=1 ⇒ BGRA. Combined with the
    // render flag FPDF_REVERSE_BYTE_ORDER, the buffer ends up
    // containing RGBA — exactly what Go's image.RGBA.Pix expects.
    const bitmap = FPDFBitmap_Create(widthPx, heightPx, 1 /* alpha */);
    if (!bitmap) {
        throw new Error(
            `pdfium: FPDFBitmap_Create(${widthPx}, ${heightPx}, alpha) failed ` +
            `— likely out of pdfium heap (page ${pageIdx} @ ${dpi} DPI, ` +
            `${widthPx * heightPx * 4} bytes needed)`);
    }

    let pagePtr = 0;
    try {
        // White background — pdfium leaves uncovered regions
        // transparent otherwise.
        FPDFBitmap_FillRect(bitmap, 0, 0, widthPx, heightPx, 0xFFFFFFFF);

        pagePtr = FPDF_LoadPage(entry.doc, pageIdx);
        if (!pagePtr) throw new Error(`pdfium: FPDF_LoadPage(${pageIdx}) failed`);

        FPDF_RenderPageBitmap(
            bitmap,
            pagePtr,
            0, 0,                 // x, y
            widthPx, heightPx,    // size_x, size_y
            0,                    // rotate (0 = no rotation)
            FPDF_ANNOT | FPDF_REVERSE_BYTE_ORDER,
        );

        const bufPtr = FPDFBitmap_GetBuffer(bitmap);
        const stride = FPDFBitmap_GetStride(bitmap);
        if (stride !== widthPx * 4) {
            // Pdfium may pad the stride. Copy row-by-row to produce a
            // packed RGBA output.
            const out = new Uint8ClampedArray(widthPx * heightPx * 4);
            const heap = pdfium.HEAPU8;
            for (let y = 0; y < heightPx; y++) {
                const srcOff = bufPtr + y * stride;
                const dstOff = y * widthPx * 4;
                out.set(heap.subarray(srcOff, srcOff + widthPx * 4), dstOff);
            }
            return { data: out, width: widthPx, height: heightPx };
        }
        // Tight stride: one Uint8ClampedArray copy of the heap window.
        // The .slice() detaches from the live heap so a later
        // malloc/free can't corrupt the data we hand to Go.
        const bytes = pdfium.HEAPU8.slice(bufPtr, bufPtr + widthPx * heightPx * 4);
        return { data: new Uint8ClampedArray(bytes.buffer), width: widthPx, height: heightPx };
    } finally {
        if (pagePtr) FPDF_ClosePage(pagePtr);
        FPDFBitmap_Destroy(bitmap);
    }
}

NAMESPACE.pdfium = {
    loadDocument, closeDocument, numPages, pageSize, renderPage,
    ready: bootstrap(),
};
