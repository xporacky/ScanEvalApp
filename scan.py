#!/usr/bin/env python3
"""
ScanEvalApp — Python scanning script.
Replaces Go scanprocessing for improved reliability.

Usage:
    python3 scan.py <pdf_path> <config_path> <question_count> <option_count>

Outputs newline-delimited JSON to stdout:
    {"type": "progress",  "message": "..."}
    {"type": "result",    "page": N, "registration_number": N, "question_number": N, "answers": "abc..."}
    {"type": "error",     "page": N, "message": "..."}
    {"type": "done",      "total": N, "failed_pages": [...]}
"""

import sys
import json
import re

import numpy as np

try:
    import fitz
except ImportError:
    print(json.dumps({"type": "error", "page": -1, "message": "chybí pymupdf: pip install pymupdf"}), flush=True)
    sys.exit(1)

try:
    import cv2
except ImportError:
    print(json.dumps({"type": "error", "page": -1, "message": "chybí opencv: pip install opencv-python"}), flush=True)
    sys.exit(1)

try:
    import pytesseract
except ImportError:
    print(json.dumps({"type": "error", "page": -1, "message": "chybí pytesseract: pip install pytesseract"}), flush=True)
    sys.exit(1)

# ── Constants (must match internal/scanprocessing/constants.go) ──────────────
PADDING                    = 10
CHECKBOX_AREA_PADDING      = -5
CHECKBOX_PADDING           = 5
BORDER_RECTANGLE_AREA_SIZE = 1_000_000
ANSWER_SQUARE_MIN_AREA     = 1800
ANSWER_SQUARE_MAX_AREA     = 4500
ANSWER_SQUARE_MIN_AREA_8   = 1800
ANSWER_SQUARE_MAX_AREA_8   = 4500
N_QUESTIONS_PER_PAGE       = 15
Q_NOT_FOUND                = -1
QR_ID_PATTERN              = re.compile(r'(?:^|[|;\s])ID\s*:\s*(\d+)')

STATE_EMPTY  = 0
STATE_X      = 1
STATE_CIRCLE = 2

# Loaded from config file
MEAN_LOW  = 0.0
MEAN_HIGH = 255.0


# ── Config ────────────────────────────────────────────────────────────────────
def load_config(path: str):
    global MEAN_LOW, MEAN_HIGH
    with open(path) as f:
        cfg = json.load(f)
    MEAN_LOW  = float(cfg["mean_intensity_x_lowest"])
    MEAN_HIGH = float(cfg["mean_intensity_x_highest"])


# ── Image helpers ─────────────────────────────────────────────────────────────
def find_contours(gray: np.ndarray):
    """Canny + morphological close + external contours (matches Go FindContours)."""
    canny = cv2.Canny(gray, 100, 200)
    k = cv2.getStructuringElement(cv2.MORPH_RECT, (3, 3))
    canny = cv2.dilate(canny, k)
    canny = cv2.erode(canny, k)
    contours, _ = cv2.findContours(canny, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_NONE)
    return contours


def find_rectangle(gray: np.ndarray, min_area: float, max_area: float = -1):
    """Returns (x1, y1, x2, y2) or None — matches Go FindRectangle."""
    for c in find_contours(gray):
        approx = cv2.approxPolyDP(c, 0.01 * cv2.arcLength(c, True), True)
        area = cv2.contourArea(approx)
        if len(approx) >= 4 and area > min_area:
            if max_area != -1 and area > max_area:
                continue
            x, y, w, h = cv2.boundingRect(approx)
            return (x, y, x + w, y + h)
    return None


def find_border_rotated_rect(gray: np.ndarray):
    """Returns cv2.minAreaRect result or None — matches Go FindBorderRotatedRectangle."""
    for c in find_contours(gray):
        approx = cv2.approxPolyDP(c, 0.01 * cv2.arcLength(c, True), True)
        if len(approx) == 4 and cv2.contourArea(approx) > BORDER_RECTANGLE_AREA_SIZE:
            return cv2.minAreaRect(approx)
    return None


def check_upside_down(gray: np.ndarray) -> bool:
    """Lower half has more contours → sheet is upside down."""
    h = gray.shape[0]
    upper = gray[:h // 2, :]
    lower = gray[h // 2:, :]
    return len(find_contours(lower)) > len(find_contours(upper))


def fix_rotation(gray: np.ndarray) -> np.ndarray:
    """Detect sheet skew angle and rotate to correct orientation."""
    rect = find_border_rotated_rect(gray)
    angle = 0.0
    if rect is not None:
        angle = rect[2] - 90.0
        if abs(angle) > 45:
            angle += 90.0
    if check_upside_down(gray):
        angle += 180.0
    if abs(angle) < 0.1:
        return gray
    h, w = gray.shape[:2]
    M = cv2.getRotationMatrix2D((w // 2, h // 2), angle, 1.0)
    return cv2.warpAffine(gray, M, (w, h))


def page_to_gray(doc, page_num: int) -> np.ndarray:
    """Extract page as grayscale numpy array at native scan resolution."""
    page = doc[page_num]
    # Try to extract embedded image (scanned PDFs store image at native DPI)
    imgs = page.get_images(full=True)
    if imgs:
        xref = imgs[0][0]
        raw = doc.extract_image(xref)
        arr = np.frombuffer(raw["image"], dtype=np.uint8)
        img = cv2.imdecode(arr, cv2.IMREAD_GRAYSCALE)
        if img is not None:
            return img
    # Fallback: render at 200 DPI
    pix = page.get_pixmap(dpi=200)
    arr = np.frombuffer(pix.samples, dtype=np.uint8)
    img = arr.reshape(pix.height, pix.width, pix.n)
    if pix.n == 4:
        img = cv2.cvtColor(img, cv2.COLOR_BGRA2GRAY)
    elif pix.n == 3:
        img = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    return img


# ── QR + OCR ──────────────────────────────────────────────────────────────────
def read_qr(gray: np.ndarray) -> str:
    """Try QR detection; retry on 2× upscaled image if first attempt fails."""
    det = cv2.QRCodeDetector()
    data, _, _ = det.detectAndDecode(gray)
    if data:
        return data
    scaled = cv2.resize(gray, (gray.shape[1] * 2, gray.shape[0] * 2), interpolation=cv2.INTER_LINEAR)
    data, _, _ = det.detectAndDecode(scaled)
    return data or ""


def reg_num_from_qr(text: str):
    if not text:
        return None
    m = QR_ID_PATTERN.search(text)
    if m:
        try:
            return int(m.group(1))
        except ValueError:
            pass
    try:
        return int(text.strip())
    except ValueError:
        return None


def reg_num_from_ocr(gray: np.ndarray):
    """OCR fallback: extract ID from top-quarter header region."""
    h, w = gray.shape[:2]
    header = gray[PADDING: h // 4 - PADDING, PADDING: w - PADDING]
    for psm in ("6", "3"):
        try:
            text = pytesseract.image_to_string(header, lang="slk", config=f"--psm {psm}")
            m = re.search(r"ID:\s*(\d+)", text)
            if m:
                return int(m.group(1))
        except Exception:
            continue
    return None


# ── Answer detection ──────────────────────────────────────────────────────────
def _area_bounds(choices: int):
    if choices == 8:
        return ANSWER_SQUARE_MIN_AREA_8, ANSWER_SQUARE_MAX_AREA_8
    return ANSWER_SQUARE_MIN_AREA, ANSWER_SQUARE_MAX_AREA


def _checkbox_pad(choices: int) -> int:
    return 4 if choices == 8 else CHECKBOX_PADDING


def get_answer(cropped: np.ndarray, row: int, choices: int) -> str:
    """Evaluate one question row — matches Go GetAnswer."""
    h, w   = cropped.shape[:2]
    col_w  = w // (choices + 1)
    row_h  = h // N_QUESTIONS_PER_PAGE
    answer = "x"
    state  = STATE_EMPTY
    pad    = _checkbox_pad(choices)
    min_a, max_a = _area_bounds(choices)

    for j in range(1, choices + 1):
        v_pad = CHECKBOX_AREA_PADDING
        if row == 0 or row == N_QUESTIONS_PER_PAGE - 1:
            v_pad = 0

        x1 = col_w * j
        y1 = v_pad + row * row_h
        x2 = col_w * (j + 1)
        y2 = (row + 1) * row_h - v_pad

        if y1 >= y2 or x1 >= x2 or y2 > h or x2 > w:
            continue

        cell = cropped[y1:y2, x1:x2]
        rect = find_rectangle(cell, min_a, max_a)

        if rect is None:
            # No border rectangle found → circle-style marking
            if state == STATE_CIRCLE:
                return "x"
            answer = chr(ord("a") + (j - 1))
            state  = STATE_CIRCLE
            continue

        rx1, ry1, rx2, ry2 = rect
        rx1 += pad; ry1 += pad; rx2 -= pad; ry2 -= pad
        if rx2 - rx1 <= 0 or ry2 - ry1 <= 0:
            continue

        mean_val = float(np.mean(cell[ry1:ry2, rx1:rx2]))
        if MEAN_LOW < mean_val < MEAN_HIGH:
            if state == STATE_EMPTY:
                answer = chr(ord("a") + (j - 1))
                state  = STATE_X
            elif state == STATE_X:
                answer = "x"

    return answer


def get_question_number(cropped: np.ndarray, row: int, choices: int) -> int:
    """OCR the question-number cell for a given row."""
    h, w  = cropped.shape[:2]
    col_w = w // (choices + 1)
    row_h = h // N_QUESTIONS_PER_PAGE

    x1 = PADDING
    y1 = PADDING + row * row_h
    x2 = col_w - PADDING
    y2 = (row + 1) * row_h - PADDING

    if y2 <= y1 or x2 <= x1:
        return Q_NOT_FOUND

    region = cropped[y1:y2, x1:x2]
    try:
        text = pytesseract.image_to_string(region, lang="slk", config="--psm 7").strip()
        return int(text)
    except Exception:
        return Q_NOT_FOUND


def crop_answers_only(gray: np.ndarray):
    """Crop to the answer-area bounding rectangle."""
    rect = find_rectangle(gray, BORDER_RECTANGLE_AREA_SIZE, -1)
    if rect is None:
        return None
    x1, y1, x2, y2 = rect
    x1 += PADDING; y1 += PADDING; x2 -= PADDING; y2 -= PADDING
    if x2 - x1 <= 0 or y2 - y1 <= 0:
        return None
    return gray[y1:y2, x1:x2]


def evaluate_answers(gray: np.ndarray, question_count: int, option_count: int):
    """
    Returns (last_question_index, answers_string) or (Q_NOT_FOUND, None).
    Matches Go EvaluateAnswers behaviour exactly.
    """
    cropped = crop_answers_only(gray)
    if cropped is None:
        return Q_NOT_FOUND, None

    answers  = []
    q_number = Q_NOT_FOUND

    for i in range(N_QUESTIONS_PER_PAGE):
        answers.append(get_answer(cropped, i, option_count))

        if q_number == Q_NOT_FOUND:
            q_number = get_question_number(cropped, i, option_count)
            continue  # don't increment yet — matches Go `continue`

        q_number += 1
        if q_number >= question_count:
            break

    if q_number == Q_NOT_FOUND:
        return Q_NOT_FOUND, None

    return q_number - 1, "".join(answers)


# ── Page processing ───────────────────────────────────────────────────────────
def process_page(doc, page_num: int, question_count: int, option_count: int) -> dict:
    gray = page_to_gray(doc, page_num)
    gray = fix_rotation(gray)

    # 1. QR code
    reg_num = reg_num_from_qr(read_qr(gray))

    # 2. OCR fallback
    if reg_num is None:
        reg_num = reg_num_from_ocr(gray)

    if reg_num is None:
        return {"type": "error", "page": page_num, "message": "Nepodarilo sa získať ID (QR ani OCR)"}

    # 3. Answer evaluation
    q_number, answers = evaluate_answers(gray, question_count, option_count)

    if not answers:
        return {"type": "error", "page": page_num, "message": "Žiadne odpovede detekované"}
    if q_number == Q_NOT_FOUND:
        return {"type": "error", "page": page_num, "message": "Číslo otázky nebolo nájdené"}

    # 4. Validate question count on page (matches Go validation in ProcessPage)
    questions_on_page = (q_number + 1) % N_QUESTIONS_PER_PAGE
    if questions_on_page == 0:
        questions_on_page = N_QUESTIONS_PER_PAGE
    is_full = questions_on_page == N_QUESTIONS_PER_PAGE
    is_last = (q_number + 1) == question_count
    if not is_full and not is_last:
        return {
            "type": "error",
            "page": page_num,
            "message": f"Neočakávaný počet otázok: posledná={q_number + 1}, expected={question_count}",
        }

    return {
        "type":                "result",
        "page":                page_num,
        "registration_number": reg_num,
        "question_number":     q_number,
        "answers":             answers,
    }


# ── Main ──────────────────────────────────────────────────────────────────────
def emit(obj: dict):
    print(json.dumps(obj, ensure_ascii=False), flush=True)


def main():
    if len(sys.argv) < 5:
        emit({"type": "error", "page": -1, "message": "Použitie: scan.py <pdf> <config> <questions> <options>"})
        sys.exit(1)

    pdf_path       = sys.argv[1]
    config_path    = sys.argv[2]
    question_count = int(sys.argv[3])
    option_count   = int(sys.argv[4])

    try:
        load_config(config_path)
    except Exception as e:
        emit({"type": "error", "page": -1, "message": f"Config chyba: {e}"})
        sys.exit(1)

    try:
        doc = fitz.open(pdf_path)
    except Exception as e:
        emit({"type": "error", "page": -1, "message": f"PDF chyba: {e}"})
        sys.exit(1)

    total    = doc.page_count
    failed   = []

    for page_num in range(total):
        result = process_page(doc, page_num, question_count, option_count)
        emit(result)
        if result["type"] != "result":
            failed.append(page_num)
        emit({"type": "progress", "message": f"Spracovaných {page_num + 1} / {total}"})

    emit({"type": "done", "total": total, "failed_pages": failed})


if __name__ == "__main__":
    main()