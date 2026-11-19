"""UI translations must follow docs/i18n-glossary.md.

Checks cmd/launcher/web/js/i18n/translations.js:
  * the canonical term for "tweak", "driver", "profile" and "update" (upd_* keys)
    is used wherever the English text uses the term;
  * pt is Brazilian Portuguese (no European-only words);
  * tr uses Turkish letters (no ASCII-folded spellings of common words);
  * nl addresses the user informally (no "u"/"uw");
  * ja contains Japanese (kana), not Chinese-only text; zh/ko contain no
    text in the wrong script.

Stems are regexes searched case-insensitively and are deliberately short so
that they survive inflection (cases, plural, verb forms).
"""
import pathlib
import re
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
TRANSLATIONS = ROOT / "cmd" / "launcher" / "web" / "js" / "i18n" / "translations.js"

LANG_RE = re.compile(r"^  (?:'([A-Za-z-]+)'|([A-Za-z]+)): \{\s*$")
KV_RE = re.compile(r"^\s+'([^']+)':\s*'((?:[^'\\]|\\.)*)',?\s*$")

LANGS = [
    "pl", "en", "bg", "cs", "da", "de", "el", "es", "et", "fi", "fr", "hr", "hu",
    "it", "lt", "lv", "nb", "nl", "pt", "ro", "ru", "sk", "sl", "sr-Latn", "sv",
    "tr", "uk", "zh", "ja", "ko",
]


def parse_translations(path=TRANSLATIONS):
    """Return {lang: {key: value}} from the `'key': 'value',` lines."""
    data = {}
    cur = None
    for line in path.read_text(encoding="utf-8").splitlines():
        m = LANG_RE.match(line)
        if m:
            cur = m.group(1) or m.group(2)
            data[cur] = {}
            continue
        if cur is None:
            continue
        m = KV_RE.match(line)
        if m:
            data[cur][m.group(1)] = m.group(2).replace("\\'", "'")
    return data


# Things that are never translated and must not count as a use of a term:
# file names (data/tweaks.json), placeholders ({count}), HTML tags.
_NOISE_RE = re.compile(
    r"\S+\.(?:json|txt|zip|exe|ps1|bat|html)\b|\{[^}]*\}|<[^>]+>|\\+[nt]")


def _clean(value):
    return _NOISE_RE.sub(" ", value)


# --------------------------------------------------------------------------
# Canonical terms (see docs/i18n-glossary.md)
# --------------------------------------------------------------------------

TWEAK = {
    "pl": r"tweak", "en": r"tweak", "bg": r"оптимизаци", "cs": r"úprav",
    "da": r"justering", "de": r"tweak", "el": r"βελτιστοπο[ιί]", "es": r"ajuste",
    "et": r"kohandu", "fi": r"sää[dt]", "fr": r"tweak", "hr": r"prilagodb",
    "hu": r"finomhangol", "it": r"tweak", "lt": r"patobulinim", "lv": r"pielāgojum",
    "nb": r"justering", "nl": r"aanpassing", "pt": r"tweak", "ro": r"ajust[aă]r",
    "ru": r"твик", "sk": r"úprav", "sl": r"prilagodit", "sr-Latn": r"izmen",
    "sv": r"justering", "tr": r"[iİ]nce ayar", "uk": r"твік", "zh": r"优化",
    "ja": r"ツイーク", "ko": r"트윅",
}

DRIVER = {
    "pl": r"sterownik", "en": r"driver", "bg": r"драйвер", "cs": r"ovladač",
    "da": r"driver", "de": r"treiber", "el": r"οδήγησ", "es": r"controlador",
    "et": r"draiver", "fi": r"ajur", "fr": r"pilote", "hr": r"upravljačk",
    "hu": r"illesztőprogram", "it": r"driver", "lt": r"tvarkykl", "lv": r"draiver",
    "nb": r"driver", "nl": r"stuurprogramma", "pt": r"driver", "ro": r"driver",
    "ru": r"драйвер", "sk": r"ovládač", "sl": r"gonilnik", "sr-Latn": r"drajver",
    "sv": r"drivrutin", "tr": r"sürücü", "uk": r"драйвер", "zh": r"驱动",
    "ja": r"ドライバー", "ko": r"드라이버",
}

PROFILE = {
    "pl": r"profil", "en": r"profile", "bg": r"профил", "cs": r"profil",
    "da": r"profil", "de": r"profil", "el": r"προφίλ", "es": r"perfil",
    "et": r"profiil", "fi": r"profiil", "fr": r"profil", "hr": r"profil",
    "hu": r"profil", "it": r"profil", "lt": r"profil", "lv": r"profil",
    "nb": r"profil", "nl": r"profiel", "pt": r"perfi[ls]", "ro": r"profil",
    "ru": r"профил", "sk": r"profil", "sl": r"profil", "sr-Latn": r"profil",
    "sv": r"profil", "tr": r"profil", "uk": r"профіл", "zh": r"配置文件",
    "ja": r"プロファイル", "ko": r"프로필",
}

UPDATE = {
    "pl": r"aktualiz", "en": r"updat", "bg": r"актуализ", "cs": r"aktualiz",
    "da": r"opdater", "de": r"update|aktualisier", "el": r"ενημ[εέ]ρ",
    "es": r"actualiz", "et": r"värskend", "fi": r"päivit", "fr": r"(?:mises?|mis|mettre) à jour",
    "hr": r"ažur", "hu": r"frissít", "it": r"aggiorn", "lt": r"naujin",
    "lv": r"atjaunin", "nb": r"oppdater", "nl": r"update|bijgewerkt|bijwerken",
    "pt": r"atualiz", "ro": r"actualiz", "ru": r"обнов", "sk": r"aktualiz",
    "sl": r"posodob", "sr-Latn": r"ažur", "sv": r"uppdater", "tr": r"güncelle",
    "uk": r"онов", "zh": r"更新", "ja": r"アップデート", "ko": r"업데이트",
}

# name -> (stems, regex on the English value, key prefix or None)
TERMS = {
    "tweak": (TWEAK, r"\btweaks?\b", None),
    "driver": (DRIVER, r"\bdrivers?\b", None),
    "profile": (PROFILE, r"\bprofiles?\b", None),
    "update": (UPDATE, r"\bupdat", "upd_"),
}

# (term, lang or "*", key) -> reason. Keep this list short.
TERM_ALLOWLIST = {
    ("tweak", "*", "workflow_step4_desc"):
        "workflow guide: 'other tweaks' is used generically next to registry "
        "changes and power plans; a paraphrase such as 'optimisations' is fine",
}


def term_problems(data):
    """{lang: [message]} for every value missing the canonical term."""
    problems = {}
    en = data["en"]
    for term, (stems, en_re, prefix) in TERMS.items():
        keys = [
            k for k, v in en.items()
            if (prefix is None or k.startswith(prefix))
            and re.search(en_re, _clean(v), re.IGNORECASE)
        ]
        for lang in LANGS:
            stem = re.compile(stems[lang], re.IGNORECASE)
            for key in keys:
                if (term, "*", key) in TERM_ALLOWLIST or (term, lang, key) in TERM_ALLOWLIST:
                    continue
                value = data.get(lang, {}).get(key)
                if value is None:
                    continue  # missing keys are i18n_test.go's job
                if not stem.search(_clean(value)):
                    problems.setdefault(lang, []).append(
                        f"[{term}] {key}: expected /{stems[lang]}/ in {value!r}")
    return problems


# --------------------------------------------------------------------------
# Portuguese: Brazilian only
# --------------------------------------------------------------------------

PT_EUROPEAN = [
    (r"\bficheiros?\b", "arquivo"),
    (r"\bguard(?:ar|e|a|ado|ados|ada|adas|ando)\b", "salvar"),
    (r"\bdefiniç(?:ão|ões)\b", "configurações"),
    (r"\becrãs?\b", "tela"),
    (r"\ba carregar\b", "carregando"),
    (r"\b(?:o teu|a tua|os teus|as tuas)\b", "seu/sua"),
    (r"\butilizador(?:es)?\b", "usuário"),
    (r"\btransferi(?:r|do|dos|da|ndo)\b", "baixar/download"),
    (r"\bsecç(?:ão|ões)\b", "seção"),
    (r"\bregistos?\b", "registro"),
    (r"\belimin(?:ar|e|a|ado|ados|ada|adas|ando)\b", "excluir"),
    # European continuous "A aplicar…" instead of the gerund "Aplicando…"
    (r"^A [a-zà-ú]+(?:ar|er|ir)\b", "gerúndio (…ando/…endo)"),
]


def pt_problems(data):
    out = []
    for key, value in data["pt"].items():
        seen = set()
        for pattern, brazilian in PT_EUROPEAN:
            flags = 0 if pattern.startswith("^") else re.IGNORECASE
            for m in re.finditer(pattern, value, flags):
                word = m.group(0).lower()
                if word in seen:
                    continue
                seen.add(word)
                out.append(f"{key}: European {m.group(0)!r} (pt-BR: {brazilian}) in {value!r}")
    return out


# --------------------------------------------------------------------------
# Turkish: diacritics
# --------------------------------------------------------------------------

# ASCII-folded spelling -> correct spelling. Every pattern only matches text
# that has lost its Turkish letters, so correctly written words never match.
TR_ASCII = [
    (r"\bklasor\w*", "klasör"),
    (r"\bsurucu\w*", "sürücü"),
    (r"\byukle\w*", "yükle"),
    (r"\bsec(?:in|ili|ilen|ildi|ilmemis|im|imi|enek\w*)?\b", "seç"),
    (r"\bdegis\w*", "değiş"),
    (r"\bdegil\b", "değil"),
    (r"\bbaslat\w*", "başlat"),
    (r"\bbasar\w*", "başarı"),
    (r"\bac\b", "aç"),
    (r"\bgoster\w*", "göster"),
    (r"\bduzenle\w*", "düzenle"),
    (r"\btum\w*", "tüm"),
    (r"\bbolum\w*", "bölüm"),
    (r"\born\b", "örn."),
    (r"\bkayit\w*", "kayıt"),
    (r"\bguc\b", "güç"),
    (r"\bdiger\w*", "diğer"),
    (r"\bbulunamadi\b", "bulunamadı"),
    (r"\bice aktar\w*|\bdisari\b|\bdisa aktar\w*", "içe/dışa aktar"),
    (r"\btasinabilir\b", "taşınabilir"),
    (r"\bcalistir\w*", "çalıştır"),
    (r"\bkaldir\w*", "kaldır"),
    (r"\byonetici\w*", "yönetici"),
    (r"\bsimdi\b", "şimdi"),
    (r"\bsurum\w*", "sürüm"),
    (r"\bayar(?:i|lari|larini|ini)\b", "ayarı/ayarları"),
    (r"\bhatasi\b", "hatası"),
    (r"\bbankasi\b", "bankası"),
    (r"\bgerektiginde\b", "gerektiğinde"),
    (r"\bInce\b", "İnce"),
]


def tr_problems(data):
    out = []
    for key, value in data["tr"].items():
        text = _clean(value)
        for pattern, correct in TR_ASCII:
            flags = re.ASCII | (0 if pattern == r"\bInce\b" else re.IGNORECASE)
            m = re.search(pattern, text, flags)
            if m:
                out.append(f"{key}: ASCII {m.group(0)!r} (should be {correct}) in {value!r}")
    return out


# --------------------------------------------------------------------------
# Dutch: informal je/jouw
# --------------------------------------------------------------------------

NL_FORMAL_RE = re.compile(r"(?<![\w'+-])(?:u|uw)(?![\w'+-])", re.IGNORECASE)


def nl_problems(data):
    out = []
    for key, value in data["nl"].items():
        m = NL_FORMAL_RE.search(_clean(value))
        if m:
            out.append(f"{key}: formal {m.group(0)!r} (use je/jouw) in {value!r}")
    return out


# --------------------------------------------------------------------------
# CJK scripts
# --------------------------------------------------------------------------

def _range(lo, hi):
    return f"{chr(lo)}-{chr(hi)}"


KANA = _range(0x3040, 0x30FF)        # hiragana + katakana
KATAKANA = _range(0x30A1, 0x30FF)
HAN = _range(0x3400, 0x9FFF)          # CJK ideographs
KANA_RE = re.compile(f"[{KANA}]")
HAN_RE = re.compile(f"[{HAN}]")
# Simplified-only characters; Japanese writes them differently
# (e.g. 个 -> 個, 选 -> 選, 优 -> 優, 设 -> 設).
ZH_ONLY_RE = re.compile("[这个们么没显从应优调预设时为发务进关开误动择载档选该请]")
# A Latin fragment sandwiched inside a katakana word, e.g. "ランcherー".
JA_MIXED_RE = re.compile(f"[{KATAKANA}][a-z]+[{KATAKANA}]")
# Japanese UI text may be short Han-only compounds (設定, 安定版, 状態: 未確認);
# a run of more than this many Han characters without any kana is Chinese.
JA_MAX_HAN_WITHOUT_KANA = 6

JA_ALLOWLIST = {
    # key -> reason (kana-free text that is correct Japanese)
}


def ja_problems(data):
    out = []
    for key, value in data["ja"].items():
        text = _clean(value)
        m = ZH_ONLY_RE.search(text)
        if m:
            out.append(f"{key}: Chinese-only character {m.group(0)!r} in {value!r}")
            continue
        if JA_MIXED_RE.search(text):
            out.append(f"{key}: Latin letters inside a katakana word in {value!r}")
            continue
        if key in JA_ALLOWLIST:
            continue
        han = len(HAN_RE.findall(text))
        if han > JA_MAX_HAN_WITHOUT_KANA and not KANA_RE.search(text):
            out.append(f"{key}: {han} Han characters and no kana (Chinese text?) in {value!r}")
    return out


def zh_problems(data):
    return [f"{k}: Japanese kana in {v!r}" for k, v in data["zh"].items() if KANA_RE.search(v)]


def ko_problems(data):
    return [f"{k}: Han/kana instead of Hangul in {v!r}" for k, v in data["ko"].items()
            if HAN_RE.search(v) or KANA_RE.search(v)]


SCRIPT_CHECKS = {"pt": pt_problems, "tr": tr_problems, "nl": nl_problems,
                 "ja": ja_problems, "zh": zh_problems, "ko": ko_problems}


def all_problems(data=None):
    """{lang: [message]} — every glossary violation, for reports."""
    data = data or parse_translations()
    problems = term_problems(data)
    for lang, check in SCRIPT_CHECKS.items():
        problems.setdefault(lang, []).extend(check(data))
    return {lang: msgs for lang, msgs in problems.items() if msgs}


class GlossaryTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.data = parse_translations()

    def test_parser_sees_all_languages(self):
        self.assertEqual(sorted(self.data), sorted(LANGS))
        for lang in LANGS:
            self.assertGreater(len(self.data[lang]), 500, lang)

    def test_stems_defined_for_every_language(self):
        for stems, _, _ in TERMS.values():
            self.assertEqual(sorted(stems), sorted(LANGS))

    def test_canonical_terms(self):
        problems = term_problems(self.data)
        for lang in LANGS:
            with self.subTest(lang=lang):
                msgs = problems.get(lang, [])
                self.assertEqual(msgs, [], f"{lang}: {len(msgs)} problem(s)\n  " + "\n  ".join(msgs))

    def test_language_specific_rules(self):
        for lang, check in SCRIPT_CHECKS.items():
            with self.subTest(lang=lang):
                msgs = check(self.data)
                self.assertEqual(msgs, [], f"{lang}: {len(msgs)} problem(s)\n  " + "\n  ".join(msgs))


if __name__ == "__main__":
    unittest.main()
