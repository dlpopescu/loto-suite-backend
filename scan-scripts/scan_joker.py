import pytesseract
import re
from PIL import Image
import sys
import json

def extract_joker_info(image_path, debug=False):
    """Extract lottery info specifically for JOKER tickets"""
    # OCR the image
    text = pytesseract.image_to_string(Image.open(image_path), lang='eng')
    lines = [line.strip().upper() for line in text.splitlines() if line.strip()]
    
    if debug:
        print("\n=== RAW OCR OUTPUT ===")
        for i, line in enumerate(lines, 1):
            print(f"{i}: {repr(line)}")
        print("======================\n")
    
    result = {'game': 'JOKER'}
    
    # Extract draw date: DT:DD/MM/YY
    result['draw_date'] = None # pyright: ignore[reportArgumentType]
    for line in lines:
        m = re.search(r'DT[:\s;]*(\d{2}[/|]\d{2}[/|]\d{2})', line)
        if m:
            result['draw_date'] = m.group(1).replace('|', '/')
            break
    
    # Extract variants: A05, B05 (max 2 variants, each with 5 main numbers + 1 joker number = 6 total)
    # First extract the 5 main numbers per variant
    variants_main = []
    seen_ids = set()
    
    for line in lines:
        # First try: Match "A05: 01,10,12,26,29" pattern
        # Handle OCR errors by allowing junk before the letter
        m = re.search(r'[^A-B]*([A-B])(?:O|0)(?:5|S)[:\s;]+([0-9\s,]+)', line)
        if m and m.group(1) not in seen_ids:
            nums = [int(n) for n in re.findall(r'\d+', m.group(2))]
            if len(nums) >= 5:
                nums = nums[-5:] if len(nums) > 5 else nums
                if len(nums) == 5 and all(1 <= n <= 45 for n in nums):
                    variants_main.append({'id': m.group(1), 'numbers': nums})
                    seen_ids.add(m.group(1))
                    if debug:
                        print(f"Pass 1: Matched variant {m.group(1)}: {nums}")
                    continue
        
        # Second try: More flexible pattern
        m2 = re.search(r'[^A-B]*([A-B])[^A-B:\d;]{0,6}[:\s;]+([0-9\s,]+)', line)
        if m2 and m2.group(1) not in seen_ids:
            nums = [int(n) for n in re.findall(r'\d+', m2.group(2))]
            if len(nums) >= 5:
                nums = nums[-5:] if len(nums) > 5 else nums
                if len(nums) == 5 and all(1 <= n <= 45 for n in nums):
                    variants_main.append({'id': m2.group(1), 'numbers': nums})
                    seen_ids.add(m2.group(1))
                    if debug:
                        print(f"Pass 2: Matched variant {m2.group(1)}: {nums}")
                    continue
        
        # Fallback: if we find exactly 5 numbers in sequence
        if len(seen_ids) < 2:
            nums = [int(n) for n in re.findall(r'\d+', line)]
            if len(nums) == 5 and all(1 <= n <= 45 for n in nums):
                for letter in 'AB':
                    if letter not in seen_ids:
                        variants_main.append({'id': letter, 'numbers': nums})
                        seen_ids.add(letter)
                        if debug:
                            print(f"Fallback matched variant {letter}: {nums}")
                        break
    
    # Extract joker numbers: A01, B01 (one joker number per variant)
    # These are typically 2-digit numbers appearing on separate lines
    joker_numbers = {}
    
    for line in lines:
        # Very flexible pattern to handle OCR errors
        # Can match: A01: 06 - QP, AGI: 06 - OF, AO1: 20, etc.
        m = re.search(r'([A-B])[^:\d]*[:\s)]+(\d{1,2})\s*[)\s]*[-\s]*[OQF][PF]', line)
        if m:
            joker_id = m.group(1)
            joker_num = int(m.group(2))
            # Joker numbers are typically < 100 and should match a variant ID
            if joker_num < 100:
                joker_numbers[joker_id] = joker_num
                if debug:
                    print(f"Matched joker for {joker_id}: {joker_num}")
    
    # Combine main numbers with joker numbers to create full variants (6 numbers each)
    variants = []
    for v in variants_main:
        variant_id = v['id']
        nums = v['numbers'].copy()
        # Append the joker number as the 6th number
        if variant_id in joker_numbers:
            nums.append(joker_numbers[variant_id])
        variants.append({'id': variant_id, 'numbers': nums})
    
    result['variants'] = variants # pyright: ignore[reportArgumentType]
    
    # Extract Noroc Plus: NOROC PLUS: XXXXXX (6 digits)
    result['noroc'] = None # pyright: ignore[reportArgumentType]
    for line in lines:
        m = re.search(r'NOROC\s+PLUS[^:]*:\s*(\d{6})', line)
        if m:
            result['noroc'] = m.group(1)
            break
    
    return result

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python ticket_joker.py <image_path> [--debug] [--json]")
        print("\nExtract lottery info from JOKER tickets")
        print("  --debug: Show raw OCR output")
        print("  --json: Output in JSON format (default: human-readable)")
        sys.exit(1)
    
    image_path = sys.argv[1]
    debug = '--debug' in sys.argv
    json_output = '--json' in sys.argv
    
    info = extract_joker_info(image_path, debug=debug)
    
    if json_output:
        # Convert to the cache format
        # Map variant letters (A, B) to 1-based numeric indices (1, 2)
        # Frontend expects 1-based IDs to match input element IDs like number-input-1-1
        letter_to_index = {'A': 1, 'B': 2}
        output = {
            "game_id": "joker",
            "game_date": info.get('draw_date', ''),
            "nume_noroc": "NOROC PLUS",
            "variante": [
                {
                    "id": letter_to_index.get(v['id'], 1), # pyright: ignore[reportArgumentType]
                    "numere": [
                        {
                            "numar": str(num),
                        }
                        for num in v['numbers'] # pyright: ignore[reportArgumentType]
                    ]
                }
                for v in info.get('variants', [])
            ],
            "noroc": {
                "numar": info.get('noroc', '')
            }
        }
        print(json.dumps(output, indent=2, ensure_ascii=False))
    else:
        # Human-readable format
        print("\n" + "="*50)
        print("JOKER TICKET SCAN RESULTS")
        print("="*50)
        print(f"Draw Date: {info.get('draw_date', 'N/A')}")
        print(f"\nVariants ({len(info.get('variants', []))} detected):")
        for v in info.get('variants', []):
            nums_str = ', '.join(map(str, v['numbers'])) # pyright: ignore[reportArgumentType]
            print(f"  {v['id']}: [{nums_str}] (last number is joker)") # pyright: ignore[reportArgumentType]
        
        print(f"\nNoroc Plus (6 digits): {info.get('noroc', 'N/A')}")
        print("="*50 + "\n")
