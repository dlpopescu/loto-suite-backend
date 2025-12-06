#!/usr/bin/env python3
"""
Quick game type detection from lottery ticket images.
Extracts only the game name to verify ticket type before full scan.
"""
import pytesseract
from PIL import Image
import sys
import json
import re

def find_ticket_bounds(img):
    """
    Find the bounding box of the white/light ticket in the image.
    Returns (left, top, right, bottom) or None if not found.
    """
    try:
        import numpy as np
        from PIL import ImageOps
        
        # Convert to grayscale
        gray = img.convert('L')
        
        # Convert to numpy array for processing
        img_array = np.array(gray)
        
        # Threshold to find bright areas (ticket is white/light)
        # Anything brighter than 150 (out of 255) is considered ticket
        threshold = 150
        bright_mask = img_array > threshold
        
        # Find rows and columns that contain bright pixels
        rows_with_ticket = np.any(bright_mask, axis=1)
        cols_with_ticket = np.any(bright_mask, axis=0)
        
        # Find the bounds
        if not np.any(rows_with_ticket) or not np.any(cols_with_ticket):
            return None
        
        top = np.argmax(rows_with_ticket)
        bottom = len(rows_with_ticket) - np.argmax(rows_with_ticket[::-1])
        left = np.argmax(cols_with_ticket)
        right = len(cols_with_ticket) - np.argmax(cols_with_ticket[::-1])
        
        # Add small margin
        margin = 10
        top = max(0, top - margin)
        left = max(0, left - margin)
        bottom = min(img.height, bottom + margin)
        right = min(img.width, right + margin)
        
        return (left, top, right, bottom)
    except Exception as e:
        print(f"Error finding ticket bounds: {e}", file=sys.stderr)
        return None

def detect_game_type(image_path, debug=False):
    """
    Detect which game type a ticket is for.
    Returns one of: '649', '540', 'joker', or None if unknown
    """
    try:
        # Open image
        img = Image.open(image_path)
        
        # Try to find and crop to just the ticket area
        bounds = find_ticket_bounds(img)
        if bounds:
            if debug:
                print(f"DEBUG: Found ticket bounds: {bounds}", file=sys.stderr)
            img = img.crop(bounds)
        else:
            if debug:
                print("DEBUG: Could not detect ticket bounds, using full image", file=sys.stderr)
        
        # Crop to top 50% where game name appears
        width, height = img.size
        top_half = img.crop((0, 0, width, int(height * 0.5)))
        
        # Preprocess image to improve OCR accuracy
        # Convert to grayscale
        top_half = top_half.convert('L')
        
        # Increase contrast and apply threshold to make text stand out
        from PIL import ImageEnhance, ImageOps
        
        # Enhance contrast
        enhancer = ImageEnhance.Contrast(top_half)
        top_half = enhancer.enhance(2.0)  # Increase contrast
        
        # Apply auto-contrast to normalize brightness
        top_half = ImageOps.autocontrast(top_half)
        
        # Run OCR on the preprocessed image
        text = pytesseract.image_to_string(top_half, lang='eng').upper()
        
        if debug:
            print(f"DEBUG: OCR Text from top 50%:", file=sys.stderr)
            print(text[:500], file=sys.stderr)  # First 500 chars
            print("=" * 50, file=sys.stderr)
        
        # Look for game identifiers - check most distinctive patterns first
        
        # JOKER - most distinctive, check first
        if re.search(r'JOKER', text):
            return 'joker'
        
        # SUPER LOTO or 5/40 or 5 DIN 40
        # The | means OR in regex
        if re.search(r'SUPER.*LOTO|5[/\s]*40|5\s+DIN\s+40', text):
            return '540'
        
        # LOTO 6/49 or 6/49 or 6 DIN 49
        # The | means OR in regex
        if re.search(r'6[/\s]*49|6\s+DIN\s+49|LOTO.*6.*49', text):
            return '649'
        
        return None
        
    except Exception as e:
        print(f"Error detecting game: {e}", file=sys.stderr)
        return None

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python detect_game.py <image_path> [--debug]")
        sys.exit(1)
    
    image_path = sys.argv[1]
    debug = '--debug' in sys.argv
    game_id = detect_game_type(image_path, debug=debug)
    
    # Output just the game_id string (or empty string if not detected)
    print(game_id if game_id else "")
