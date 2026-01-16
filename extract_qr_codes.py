#!/usr/bin/env python3
"""
Script to extract QR codes from scanned PDF and map them to page numbers.
This helps identify which student ID belongs to which page.
Uses OpenCV which should already be available.
"""

import sys
import subprocess
import tempfile
import os
import cv2

def extract_qr_from_page(pdf_path, page_num):
    """Extract QR code from a specific PDF page using pdftoppm and OpenCV."""
    # Convert PDF page to PNG using pdftoppm
    with tempfile.TemporaryDirectory() as tmpdir:
        output_prefix = os.path.join(tmpdir, "page")
        cmd = [
            "pdftoppm",
            "-png",
            "-f", str(page_num),
            "-l", str(page_num),
            "-r", "300",
            pdf_path,
            output_prefix
        ]
        
        try:
            subprocess.run(cmd, check=True, capture_output=True)
            
            # Find the generated PNG file
            png_file = f"{output_prefix}-{page_num}.png"
            if not os.path.exists(png_file):
                # Try alternative naming
                files = os.listdir(tmpdir)
                if files:
                    png_file = os.path.join(tmpdir, files[0])
                else:
                    return None
            
            # Read image with OpenCV
            img = cv2.imread(png_file)
            if img is None:
                return None
            
            # Detect QR code
            qr_detector = cv2.QRCodeDetector()
            data, bbox, _ = qr_detector.detectAndDecode(img)
            
            if data:
                return data
            
        except subprocess.CalledProcessError:
            pass
        except Exception as e:
            print(f"Error on page {page_num}: {e}")
    
    return None

def get_pdf_page_count(pdf_path):
    """Get the number of pages in a PDF."""
    try:
        result = subprocess.run(
            ["pdfinfo", pdf_path],
            capture_output=True,
            text=True,
            check=True
        )
        for line in result.stdout.split('\n'):
            if line.startswith('Pages:'):
                return int(line.split(':')[1].strip())
    except:
        pass
    return 0

def main(pdf_path):
    """Extract all QR codes from PDF and print mapping."""
    print(f"Analyzing PDF: {pdf_path}\n")
    print("=" * 60)
    
    page_count = get_pdf_page_count(pdf_path)
    if page_count == 0:
        print("Error: Could not determine PDF page count")
        return
    
    print(f"Total pages: {page_count}\n")
    
    qr_mapping = {}
    
    for page_num in range(1, page_count + 1):
        qr_code = extract_qr_from_page(pdf_path, page_num)
        
        if qr_code:
            if qr_code not in qr_mapping:
                qr_mapping[qr_code] = []
            qr_mapping[qr_code].append(page_num)
            print(f"Page {page_num:3d}: QR Code = {qr_code}")
        else:
            print(f"Page {page_num:3d}: No QR code found")
    
    print("\n" + "=" * 60)
    print("\nSummary - QR Code to Pages mapping:")
    print("-" * 60)
    
    if qr_mapping:
        for qr_code in sorted(qr_mapping.keys(), key=lambda x: int(x) if x.isdigit() else 0):
            pages = qr_mapping[qr_code]
            print(f"QR Code {qr_code:4s} -> Pages: {', '.join(map(str, pages))}")
    else:
        print("No QR codes found in the PDF")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 extract_qr_codes.py <pdf_file>")
        sys.exit(1)
    
    main(sys.argv[1])
