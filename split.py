import csv
import os
from collections import defaultdict
from datetime import datetime

INPUT_FILE = "zoznam_po_miestnostiach.csv"
OUTPUT_DIR = "output_test"

os.makedirs(OUTPUT_DIR, exist_ok=True)

rows_by_date = defaultdict(list)

with open(INPUT_FILE, newline='', encoding='utf-8-sig') as f:
    reader = csv.DictReader(f, delimiter=';')
    for row in reader:
        datum = row['Dátum'].strip()
        miestnost = row['Miestnosť'].strip()[:5].upper()

        narodenie_raw = row['Narodenie'].strip()
        try:
            narodenie = datetime.strptime(narodenie_raw, '%d. %m. %Y').strftime('%Y-%m-%d')
        except ValueError:
            narodenie = narodenie_raw  # nechaj pôvodné ak sa nepodarí parsovať

        filtered = {
            'Meno':                  row['Meno'].strip(),
            'Priezvisko':            row['Priezvisko'].strip(),
            'Dátum narodenia':       narodenie,
            'Registračné číslo': row['Reg. č.'].strip().replace('/', ''),
            'Miestnosť':             miestnost,
        }
        rows_by_date[datum].append(filtered)

OUTPUT_FIELDS = ['Meno', 'Priezvisko', 'Dátum narodenia', 'Registračné číslo', 'Miestnosť']

for datum, rows in rows_by_date.items():
    safe_name = datum.replace('.', '-').replace('/', '-').strip()
    out_path = os.path.join(OUTPUT_DIR, f"{safe_name}.csv")
    with open(out_path, 'w', newline='', encoding='utf-8-sig') as f:
        writer = csv.DictWriter(f, fieldnames=OUTPUT_FIELDS, delimiter=',')
        writer.writeheader()
        writer.writerows(rows)
    print(f"✓ {out_path}  ({len(rows)} riadkov)")

print(f"\nHotovo – vytvorených {len(rows_by_date)} súborov.")