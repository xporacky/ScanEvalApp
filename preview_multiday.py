"""
Náhľad multi-termín PDF generovania.

Použitie:
    python preview_multiday.py <cesta_k_csv>

Vypíše ako by sa študenti zoskupili do PDF súborov (1 súbor = 1 deň),
zoradení podľa: Čas → Miestnosť → Priezvisko.
"""

import csv
import sys
from collections import defaultdict
from datetime import datetime


def parse_date(raw: str) -> datetime | None:
    for fmt in ("%d.%m.%Y", "%-d.%-m.%Y", "%d. %m. %Y"):
        try:
            return datetime.strptime(raw.strip(), fmt)
        except ValueError:
            continue
    # Skúsime aj formát bez leading zeros
    parts = raw.strip().split(".")
    if len(parts) == 3:
        try:
            return datetime(int(parts[2]), int(parts[1]), int(parts[0]))
        except ValueError:
            pass
    return None


def main():
    if len(sys.argv) < 2:
        print("Použitie: python preview_multiday.py <cesta_k_csv>")
        sys.exit(1)

    csv_path = sys.argv[1]

    try:
        with open(csv_path, newline="", encoding="utf-8-sig") as f:
            reader = csv.reader(f, delimiter=";")
            rows = list(reader)
    except FileNotFoundError:
        print(f"Chyba: Súbor '{csv_path}' neexistuje.")
        sys.exit(1)

    if not rows:
        print("CSV je prázdne.")
        sys.exit(1)

    header = rows[0]
    print(f"Stĺpce: {header}\n")

    # Stĺpce: 0=Por. 1=Priezvisko 2=Meno 3=Narodenie 4=Reg.č. 5=E-mail
    #          6=Termín 7=Čas 8=Dátum 9=Miestnosť 10=Program
    errors = []
    days: dict[str, list[dict]] = defaultdict(list)

    for i, row in enumerate(rows[1:], start=2):
        if len(row) < 10:
            errors.append(f"Riadok {i}: príliš málo stĺpcov ({len(row)})")
            continue

        surname = row[1].strip()
        name = row[2].strip()
        reg_raw = row[4].strip().replace("/", "")
        exam_time = row[7].strip()
        datum_raw = row[8].strip()
        room_raw = row[9].strip()[:5].upper()

        exam_date = parse_date(datum_raw)
        if not exam_date:
            errors.append(f"Riadok {i} ({surname} {name}): nepodarilo sa parsovať dátum '{datum_raw}'")
            continue

        day_key = exam_date.strftime("%Y-%m-%d")
        days[day_key].append({
            "surname": surname,
            "name": name,
            "reg": reg_raw,
            "time": exam_time,
            "room": room_raw,
            "date": exam_date,
        })

    if errors:
        print("=== CHYBY PRI PARSOVANÍ ===")
        for e in errors:
            print(f"  ⚠  {e}")
        print()

    if not days:
        print("Žiadni študenti neboli úspešne načítaní.")
        sys.exit(1)

    sorted_days = sorted(days.keys())
    total_students = 0

    print(f"=== PREHĽAD: {len(sorted_days)} PDF súbor(ov) by sa vygenerovalo ===\n")

    for day_key in sorted_days:
        students = days[day_key]
        # Zoraď: Čas → Miestnosť → Priezvisko
        students.sort(key=lambda s: (s["time"], s["room"], s["surname"]))

        date_label = students[0]["date"].strftime("%d.%m.%Y")
        pdf_name = f"<nazov_testu>_{day_key}.pdf"

        print(f"  📄 {pdf_name}  ({date_label}, {len(students)} študentov)")
        print(f"  {'#':<5} {'Čas':<8} {'Miestnosť':<10} {'Priezvisko':<20} {'Meno':<15} {'Reg.č.'}")
        print(f"  {'-'*70}")

        for idx, s in enumerate(students, start=1):
            print(f"  {idx:<5} {s['time']:<8} {s['room']:<10} {s['surname']:<20} {s['name']:<15} {s['reg']}")

        print()
        total_students += len(students)

    print(f"Spolu: {total_students} študentov v {len(sorted_days)} dňoch.")


if __name__ == "__main__":
    main()
