#!/usr/bin/env python3
"""Generuje pliki HTML KnowledgeBase z plików *-dane.js

Uzycie:
    python generate_kb.py                 # wszystkie pliki dane/*-dane.js
    python generate_kb.py komiksy gry     # tylko wybrane klucze
    python generate_kb.py --check         # tylko test parsowania, nic nie zapisuje
    python generate_kb.py --out=DIR       # zapis do DIR zamiast do data/KnowledgeBase/

Zrodlo i wynik: data/KnowledgeBase/ (dane uzytkownika obok CoreDeck.exe,
ignorowane przez Git, nieosadzane w EXE).
"""

import json5 as json, os, re, html, sys

TEMPLATE = '''<!DOCTYPE html>
<html lang="pl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{title} - Lokalna Baza Wiedzy</title>
    <link rel="stylesheet" href="style.css">
    <style>
        .back-btn {{ position:fixed; top:20px; left:20px; z-index:9999; text-decoration:none; background:#161b22; color:#fff; padding:12px 24px; border-radius:30px; font-weight:600; border:1px solid rgba(255,255,255,0.2); box-shadow:0 4px 12px rgba(0,0,0,0.5); transition:all 0.3s; font-family:sans-serif; font-size:0.9rem; }}
        .back-btn:hover {{ background:#1f6feb; transform:translateY(-2px); }}
    </style>
</head>
<body>
    <a href="index.html" class="back-btn">← Wróć do bazy</a>
    <header class="main-header">
        <h1>{title}</h1>
        <p>Lokalna kolekcja linków i zasobów.</p>
    </header>
    <main class="content-area" style="margin-top:80px; padding:2rem; max-width:1200px; margin-left:auto; margin-right:auto;">
        {featured_html}
        {sections_html}
    </main>
</body>
</html>'''


def _find_object(js_content):
    """Znajduje tekst obiektu przypisanego do pierwszej deklaracji const/let/var.

    Zwraca None gdy nie znaleziono. Licznik nawiasow pomija zawartosc
    stringow i komentarzy, wiec dziala niezaleznie od koncowego srednika.
    """
    m = re.search(r'\b(?:const|let|var)\s+[A-Za-z_$][\w$]*\s*=\s*\{', js_content)
    if not m:
        return None
    start = m.end() - 1  # pozycja '{'
    depth = 0
    i = start
    n = len(js_content)
    while i < n:
        c = js_content[i]
        if c in '"\'`':
            quote = c
            i += 1
            while i < n:
                if js_content[i] == '\\':
                    i += 2
                    continue
                if js_content[i] == quote:
                    break
                i += 1
        elif c == '/' and i + 1 < n and js_content[i + 1] == '/':
            i = js_content.find('\n', i)
            if i == -1:
                break
        elif c == '/' and i + 1 < n and js_content[i + 1] == '*':
            end = js_content.find('*/', i + 2)
            i = end + 1 if end != -1 else n
        elif c == '{':
            depth += 1
        elif c == '}':
            depth -= 1
            if depth == 0:
                return js_content[start:i + 1]
        i += 1
    return None


def extract_data(js_content):
    """Wyciaga obiekt danych z pliku JS (dowolna nazwa zmiennej, srednik opcjonalny)."""
    text = _find_object(js_content)
    if text is None:
        return None
    try:
        return json.loads(text)
    except ValueError as e:
        print(f"  WARN: JSON parse error: {e}")
        return None


def _link_name(l):
    return l.get("name") or l.get("title") or l.get("url", "")


def _render_links(links):
    out = ""
    for l in links:
        if not isinstance(l, dict) or not l.get("url"):
            continue
        tag = f'<span style="color:#8b949e;font-size:0.75rem;">({html.escape(l.get("tag",""))})</span> ' if l.get("tag") else ""
        star = "⭐ " if l.get("isStarred") else ""
        out += (f'<li style="margin:6px 0;">{star}<a href="{html.escape(l["url"])}" target="_blank" '
                f'style="color:#58a6ff;">{html.escape(_link_name(l))}</a> {tag}</li>')
    return out


def _render_featured(featured):
    """featured moze byc lista linkow albo obiektem {title, links}."""
    if isinstance(featured, dict):
        title = featured.get("title") or "⭐ Wyróżnione"
        links = featured.get("links", [])
    elif isinstance(featured, list):
        title = "⭐ Wyróżnione"
        links = featured
    else:
        return ""
    items = ""
    for f in links:
        if not isinstance(f, dict) or not f.get("url"):
            continue
        badge = (f'<span style="background:#1f6feb;color:#fff;padding:2px 8px;border-radius:4px;'
                 f'font-size:0.7rem;margin-left:8px;">{html.escape(f.get("badge",""))}</span>') if f.get("badge") else ""
        items += (f'<li style="margin:12px 0;padding:12px;background:rgba(255,255,255,0.03);border-radius:8px;">'
                  f'<a href="{html.escape(f["url"])}" target="_blank" style="color:#58a6ff;font-weight:600;'
                  f'font-size:1.05rem;">{html.escape(_link_name(f))}</a>{badge}<br>'
                  f'<span style="color:#8b949e;font-size:0.9rem;">{html.escape(f.get("description",""))}</span></li>')
    if not items:
        return ""
    return (f'<section style="margin-bottom:2rem;"><h2 style="color:#f0f6fc;border-bottom:1px solid #30363d;'
            f'padding-bottom:8px;">{html.escape(title)}</h2><ul style="list-style:none;padding:0;">{items}</ul></section>')


def _section(title, body):
    return (f'<section style="margin-bottom:2rem;"><h2 style="color:#f0f6fc;border-bottom:1px solid #30363d;'
            f'padding-bottom:8px;">{html.escape(title)}</h2>{body}</section>')


def _subsection(title, links_html):
    return (f'<div style="margin-bottom:1rem;"><h4 style="color:#c9d1d9;margin-bottom:0.5rem;">'
            f'{html.escape(title)}</h4><ul style="list-style:none;padding-left:1rem;margin:0;">{links_html}</ul></div>')


def _render_group_body(val, depth=0):
    """Renderuje zawartosc dodatkowej grupy: dict moze zawierac listy linkow
    albo kolejne zagniezdzone dicty (np. {"categories": {desktop: [...]}}).
    Klucze czysto opisowe ('title', 'toc_title', 'id') sa pomijane."""
    if depth > 4 or not isinstance(val, dict):
        return ""
    body = ""
    for subkey, sub in val.items():
        if subkey in ('title', 'toc_title', 'id', 'description'):
            continue
        if isinstance(sub, list):
            rendered = _render_links(sub)
            if rendered:
                body += _subsection(str(subkey), rendered)
        elif isinstance(sub, dict):
            # zagniezdzona grupa - jesli sama zawiera linki, renderuj jako liste
            if isinstance(sub.get('url'), str):
                rendered = _render_links([sub])
                if rendered:
                    body += f'<ul style="list-style:none;padding-left:1rem;margin:0;">{rendered}</ul>'
                continue
            body += _render_group_body(sub, depth + 1)
    return body


def generate_html(data, title):
    """Generuje HTML z danych."""
    featured_html = _render_featured(data.get('featured')) if isinstance(data, dict) else ""

    sections_html = ""
    for s in (data.get('sections') or []) if isinstance(data, dict) else []:
        if not isinstance(s, dict):
            continue
        body = ""
        for sub in s.get('subsections', []) or []:
            if isinstance(sub, dict):
                body += _subsection(sub.get("title", ""), _render_links(sub.get('links', []) or []))
        direct = _render_links(s.get('links', []) or [])
        if direct:
            body += f'<ul style="list-style:none;padding-left:1rem;margin:0;">{direct}</ul>'
        sections_html += _section(s.get("title", ""), body)

    # Dodatkowe grupy najwyzszego poziomu (np. "apps": {desktop: [...], mobile: [...]}
    # albo zagniezdzone: {"title": ..., "categories": {desktop: [...], ...}})
    if isinstance(data, dict):
        for key, val in data.items():
            if key in ('featured', 'sections') or not isinstance(val, dict):
                continue
            body = _render_group_body(val)
            if body:
                group_title = val.get('title') if isinstance(val.get('title'), str) else str(key).title()
                sections_html += _section(group_title, body)

    return TEMPLATE.format(title=html.escape(title), featured_html=featured_html, sections_html=sections_html)


TITLES = {
    'ai': '🤖 AI',
    'edukacja': '📚 Edukacja',
    'ebooki': '📖 eBooki',
    'filmy': '🎬 Filmy i Seriale',
    'gry': '🎮 Gry',
    'komiksy': '💬 Komiksy',
    'linux-macos': '🐧 Linux i macOS',
    'livetv': '📺 Live TV',
    'media-spolecznosciowe': '👥 Media Społecznościowe',
    'miedzynarodowe': '🌍 Międzynarodowe',
    'muzyka': '🎵 Muzyka',
    'narzedzia-graficzne': '🎨 Narzędzia Graficzne',
    'narzedzia-plikow': '📁 Narzędzia Plikowe',
    'narzedzia-systemowe': '⚙️ Narzędzia Systemowe',
    'narzedzia-tekstowe': '📝 Narzędzia Tekstowe',
    'pobieranie': '⬇️ Pobieranie',
    'poradnik-gry': '🎯 Poradnik Gry',
    'pozostale': '📦 Pozostałe',
    'privacy': '🔒 Prywatność',
    'programowanie': '💻 Programowanie',
    'raport_linkow': '📊 Raport Linków',
    'rozszerzenia': '🧩 Rozszerzenia',
    'sport': '⚽ Sport',
    'telefony': '📱 Telefony',
    'torrenty': '🏴‍☠️ Torrenty',
    'wideo': '🎥 Wideo',
    'wschodni-skarbiec': '👤 Wschodni Skarbiec',
}


def main(argv):
    check_only = '--check' in argv
    only = [a for a in argv if not a.startswith('-')]
    out_dir = None
    for a in argv:
        if a.startswith('--out='):
            out_dir = a[len('--out='):]

    root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    # The KnowledgeBase is user data: it lives in data/KnowledgeBase next to
    # CoreDeck.exe (served at /KnowledgeBase/), never inside the embedded UI,
    # so it is not compiled into the public exe or committed.
    kb_dir = os.path.join(root_dir, 'data', 'KnowledgeBase')
    dane_dir = os.path.join(kb_dir, 'dane')
    write_dir = out_dir or kb_dir
    if out_dir and not check_only:
        os.makedirs(out_dir, exist_ok=True)

    if not os.path.exists(dane_dir):
        print(f"ERROR: {dane_dir} not found")
        return 2

    ok = 0
    fail = 0
    for filename in sorted(os.listdir(dane_dir)):
        if not filename.endswith('-dane.js'):
            continue
        key = filename[:-len('-dane.js')]
        if only and key not in only:
            continue
        title = TITLES.get(key, key.replace('-', ' ').title())
        html_file = os.path.join(write_dir, f'{key}.html')

        with open(os.path.join(dane_dir, filename), 'r', encoding='utf-8') as f:
            data = extract_data(f.read())
        if not isinstance(data, dict):
            print(f"  FAIL: {filename} - cannot parse")
            fail += 1
            continue

        if check_only:
            print(f"  OK (parse) {filename}: keys={list(data.keys())}")
            ok += 1
            continue

        html_content = generate_html(data, title)
        with open(html_file, 'w', encoding='utf-8') as f:
            f.write(html_content)
        print(f"  OK {key}.html")
        ok += 1

    if only:
        missing = [k for k in only if not os.path.exists(os.path.join(dane_dir, f'{k}-dane.js'))]
        for k in missing:
            print(f"  FAIL: {k}-dane.js not found")
            fail += 1

    print(f"\nDone: {ok} {'parsed' if check_only else 'generated'}, {fail} failed")
    return 1 if fail else 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
