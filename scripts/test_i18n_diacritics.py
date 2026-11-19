"""UI texts must be written with the letters of their language.

Catches translations that were typed or machine-converted without diacritics
("ASCII-folded": pl "Usun plik", de "Schlussel", tr "Klasoru sec"), Russian
letters in Ukrainian/Bulgarian, Polish letters leaking into other languages,
cedilla instead of comma in Romanian, tilde instead of double acute in
Hungarian and Greek words without a stress mark.

Checks cmd/launcher/web/js/i18n/translations.js and tweak-translations.js:

  * share       - in every Latin-script language that has its own letters,
                  most longer texts contain at least one of them (a language
                  folded to ASCII drops to almost zero);
  * folded      - typical folded misspellings per language (words that do not
                  exist without their diacritics);
  * letters     - letters that do not belong to the language or its script;
  * greek_tonos - Greek words of several syllables carry a stress mark.

Run with the other script tests (`python -m unittest discover -s scripts`),
or `python scripts/test_i18n_diacritics.py report` for a plain listing.
"""
import pathlib
import re
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from test_i18n_glossary import LANGS, parse_translations  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parent.parent
TWEAKS = ROOT / "cmd" / "launcher" / "web" / "js" / "i18n" / "tweak-translations.js"

TW_LANG_RE = re.compile(r"^    (?:'([A-Za-z-]+)'|([A-Za-z]+)): \{\s*$")
TW_ENTRY_RE = re.compile(
    r"^        ([a-z0-9_]+): \{ name: '((?:[^'\\]|\\.)*)', description: '((?:[^'\\]|\\.)*)' \},?\s*$")


def parse_tweaks(path=TWEAKS):
    """{lang: {"<tweak_id>.name" / "<tweak_id>.description": text}}."""
    data, cur = {}, None
    for line in path.read_text(encoding="utf-8").splitlines():
        m = TW_LANG_RE.match(line)
        if m:
            cur = m.group(1) or m.group(2)
            data[cur] = {}
            continue
        m = TW_ENTRY_RE.match(line) if cur else None
        if m:
            data[cur][m.group(1) + ".name"] = m.group(2).replace("\\'", "'")
            data[cur][m.group(1) + ".description"] = m.group(3).replace("\\'", "'")
    return data


def corpus():
    """{lang: {key: text}}: UI texts plus tweak texts ("tweak:<id>.<field>")."""
    ui = parse_translations()
    tw = parse_tweaks()
    out = {}
    for lang in LANGS:
        texts = dict(ui.get(lang, {}))
        for k, v in tw.get(lang, {}).items():
            texts["tweak:" + k] = v
        out[lang] = texts
    return out


# Never-translated tokens: placeholders, tags, escapes, URLs, file names,
# anything with digits and CamelCase product names.
_NOISE_RE = re.compile(
    r"\{[^}]*\}|<[^>]+>|\\[nt]|https?://\S+|\S+\.(?:json|txt|zip|exe|ps1|bat|html|msi|dll|sys|inf)\b"
    r"|\b\w*\d\w*\b|\b[A-Za-z]*[a-z][A-Z]\w*\b")
WORD_RE = re.compile(r"[^\W\d_]+", re.UNICODE)


def words(text):
    return WORD_RE.findall(_NOISE_RE.sub(" ", text))


# Letters a Latin-script language has beyond ASCII (lower case).
SPECIAL = {
    "pl": "ąćęłńóśźż", "cs": "áčďéěíňóřšťúůýž", "sk": "áäčďéíĺľňóôŕšťúýž",
    "hu": "áéíóöőúüű", "ro": "ăâîșț", "tr": "çğıöşü", "de": "äöüß",
    "fr": "àâçéèêëîïôûùüÿœ", "es": "áéíñóúü", "it": "àèéìíòóù", "pt": "áâãàçéêíóôõú",
    "hr": "čćđšž", "sl": "čšž", "sr-Latn": "čćđšž", "lt": "ąčęėįšųūž",
    "lv": "āčēģīķļņšūž", "et": "äöõüšž", "da": "æøå", "nb": "æøå", "sv": "åäö", "fi": "äöå",
}

# share: minimum fraction of texts with at least SHARE_MIN_WORDS words that
# contain one of the SPECIAL letters. Set to roughly half of what the
# translations had when this test was written (November 2026: pl 87%, cs 100%, sk 100%, hu 99% ... it 39%).
SHARE_MIN_WORDS = 6
SHARE_MIN = {
    "pl": 0.40, "cs": 0.50, "sk": 0.45, "hu": 0.45, "ro": 0.45, "tr": 0.45, "de": 0.25,
    "fr": 0.40, "es": 0.30, "it": 0.15, "pt": 0.40, "hr": 0.35, "sl": 0.25, "sr-Latn": 0.40,
    "lt": 0.45, "lv": 0.45, "et": 0.40, "da": 0.30, "nb": 0.30, "sv": 0.40, "fi": 0.35,
}

# folded: ASCII spellings that are not words of the language (each one lost
# its diacritics). Lower case; matched as whole words, case-insensitively.
FOLDED = {
    "cs": (
        "nastaveni slozka slozku slozky slozce ovladac ovladace ovladacu stahnout stahovani ulozit "
        "zrusit prohlizec spravce uzivatel uzivatele uzivatelu pocitac pocitace pocitaci jeste "
        "vsechny vsechno spoustec instalator "
    ),
    "da": (
        "vaelg abn abner hojre storrelse aendr aendre aendringer forsog nodvendig tilfoj tilfojet "
        "sogning "
    ),
    "de": (
        "ubersicht loschen geloscht auswahlen ausgewahlt konnen mussen ausfuhren ausgefuhrt "
        "schaltflache datentrager hinzufugen hinzugefugt verfugbar gultig ungultig uberprufen "
        "offnen geoffnet moglich schlussel zuruck zurucksetzen anderungen spater wahrend fur "
        "benotigt "
    ),
    "es": (
        "configuracion actualizacion instalacion descripcion informacion ubicacion seleccion opcion "
        "sesion conexion funcion direccion despues tambien ademas busqueda aplicacion "
    ),
    "et": (
        "joudlus jouldus kaivita kaivitaja kaivitamine klopsa koik korval kulgpaneel labipaistvus "
        "loikelauale loikelaud lopetatud lulita lulitab maara maarab markeruut markus molemad moni "
        "naita napunaited noua noutav poordumatu reziim soltub susteem susteemi taaskaivita "
        "taaskaivitamine tooriistad tooriistakeskus tuhista tuhjenda tulemuur tuubi tuup uksikasjad "
        "uksikasjalik ulal unereziim valja valjalase valjas varskenda varskendus varskendused "
        "varskenduskanal voti votme "
    ),
    "fi": (
        "paivitys paivitykset paivita kaynnista kaynnistin kayttaja kayttajat hyvaksy tyokalut "
        "kaytossa kaytto lisaa "
    ),
    "fr": (
        "parametres telecharger telechargement selectionner selectionnez reussi echec echoue "
        "systeme fenetre etape deja acces redemarrer redemarrage repertoire categorie verifier "
        "donnees desactiver desactive reessayer bibliotheque memoire securite "
    ),
    "hr": (
        "azuriraj azuriranja azuriranje iskljuceno iskljuci iskljucuje izbrisi kasnjenje kljuc "
        "korisnicki medjuspremnik meduspremnik moze mrezno nesto nista obicne ocisti odredisna "
        "osvjezi pogreska pogreske pokretac pokretaca preskoci pricekajte proizvodac proizvodaca "
        "pronadena pronadjeno racun racunala racunalo racunalu rucno sazimanje sifriranje sprjecava "
        "sucelje sucelju trenutacna ukljuceno ukljuci upravljacki upravljackih upravljackog uredaj "
        "uredaja uredivac uspjesno vec zakljucani zavrseno zelite znacajke "
    ),
    "hu": (
        "adatvedelem alkalmazas alkalmazasok beallitas beallitasok bekapcsolasa bezaras biztonsagi "
        "eleresi ellenorizze ellenorzes eltavolitas eltavolitasa ertesites ervenytelen fajl fajlok "
        "fajlt finomhangolas finomhangolasok frissites frissitesek frissitesi hordozhato "
        "illesztoprogram illesztoprogramok inditas indito jogosultsag katalogus kepernyo kereses "
        "kijeloles kikapcsolasa kivalasztas letoltes megnyitasa megse modositas rendszergazdakent "
        "szamitogep talca telepites telepitese telepito torles tudasbazis ujrainditas valasszon "
        "verzio "
    ),
    "it": (
        "piu puo gia perche funzionalita attivita velocita qualita potra verra dovra modalita "
        "compatibilita possibilita priorita "
    ),
    "lt": (
        "atsaukta atsaukti atsisiunciamas atsisiusti busena idiegta idiegti ijungia ijungti ikeliama "
        "ikelti irankiai irankiu irasas irasyta irasyti isjungia isjungta isjungtas iskarpine "
        "issaugoti istrinti isvalyti ivesties kategoriju klaidu nesiojamasis nesiojamoji paieska "
        "paleidziama paleidziant pasalinta pasalinti pazymekite pazymeta pazymeti perziura perziureti "
        "pranesimai programu sarasas saraso slaptazodis tvarkykliu uzdaryti uzduociu uzklausa "
        "uzrakinimo zaidimai zaidimu ziniu zurnalas zymos "
    ),
    "lv": (
        "aizvert aparatura apstiprinat atinstalet atjauninajumi atjauninajums atjauninajumu "
        "atjauninasana atjauninat atlasit atvert drosiba dzesana dzest energijas ieslegt iestatijumi "
        "iestatijumu informacija instalesana instalet instaletajs izslegt izveleties izvelieties "
        "jaunaka kluda kludas konfiguracija kopet lejupielade lejupieladet lietotnu meklesana meklet "
        "nonemsana nonemt palaidejam palaidejs paplasinajumi parbaudiet parbaudit parlukot parluks "
        "pielagojumi pielagojums portativa privatums rediget restartejiet restartesana restartet "
        "saglabasana saglabat sistemas speles tiesibas veiktspeja zinasanu "
    ),
    "nb": (
        "apne apner hoyre storrelse nodvendig forsok sokeresultat "
    ),
    "pl": (
        "bezpieczenstwo blad bledu jezyk jezyka juz kanal narzedzi narzedzia odinstalowac odswiez "
        "otworz plikow pobrac podglad pomin pozniej programow prywatnosc przegladarka przelacz "
        "przywroc rowniez sciezka sciezke sciezki sie sprawdz sterownikow szczegoly tweakow udalo "
        "uprawnien uruchomic uslugi ustawien usun usuniecie usuniete uzytkownika wiecej wlasne wroc "
        "wybrac wyczysc wydajnosc wyszukiwarke zainstalowac zakladka zakladki zarzadzanie zastosowac "
        "zmien zrodlo zrodlowy "
    ),
    "pt": (
        "configuracoes configuracao atualizacao atualizacoes instalacao descricao informacao "
        "informacoes opcao opcoes sessao versao aplicacao conexao extensao funcao voce nao entao "
        "tambem seguranca acao acoes selecao usuario usuarios possivel impossivel disponivel "
        "indisponivel historico botao "
    ),
    "ro": (
        "actiuni activeaza actualizari adauga adaugati afiseaza ajustari ajustarile anuleaza "
        "aplicatia aplicatie aplicatii atentie cauta cautare confidentialitate copiaza cunostinte "
        "descarcare dezactiveaza doriti dumneavoastra editeaza fara fisier fisiere fisierele fisierul "
        "functii inapoi incercati inchide informatii inregistrare instalati instaleaza intarziere "
        "interfata notificari optiuni performanta reporniti retea salveaza sectiune selectati setari "
        "sterge stergere stergerea tinta verificati "
    ),
    "sk": (
        "aktualizacia aktualizacie aktualizacii archiv databaza hardver hardveru hladat instalacia "
        "instalator katalog katalogu kategoria kategorie kopirovat nacitava nainstalovat napajania "
        "napajenia nastroje nastrojov odinstalovat odstranit otvorit ovladac ovladace ovladacov "
        "pocitac prehliadac priecinka priecinky priecinok rezim softver softveru spravca spravcu "
        "spustac spustit stiahnut subor suborov suboru subory udaje ulozit uprava upravy vyhladavanie "
        "vymazat zaloha zalohy zavriet zrusit "
    ),
    "sl": (
        "crke graficne graficnih izbrisi izbrisite izracunavanje izvazanje izvozeni kljuc kljucnimi "
        "koncano moznosti nalozeno namescanje namescena namesceno napacna napacno nicesar obicajne "
        "obstojece odlozisca odlozisce omogoci onemogoci osvezi pocisti pocistite podcrtaje pokazi "
        "pomoc posiljanje preklici preskoceno preskoci racunalnik racunalnika racunalniku razlicica "
        "razlicice razlicico sifriranje skrbniske skrbniskimi stevke tezave uspesno uvazanje zacni "
        "zazeni zazenite zelite zvocna "
    ),
    "sr-Latn": (
        "azuriraj azuriranja azuriranje cuvanje greska greske iskljuceno iskljucuje izbrisi kasnjenje "
        "kljuc koriscenje korisceno korisnicki moze nesto nista obavestenja obrisi osvezi otkazi "
        "podesavanja podesavanje podesavanjima podrska pokretac pregledac preskoci prilagodjene "
        "proizvodac proizvodjac proizvodjaca pronadena pronadjen pronadjeni racunar racunara racunaru "
        "resenje rucno sacekajte sacuvaj sacuvano sifrovanje sprecava stampac ukljuceno ukljucuje "
        "uspesno vec zakljucani zavrseno zelite zvucna "
    ),
    "sv": (
        "installningar valj hamta anvandare oppna oppnar forsok sokning andringar tillganglig "
        "nodvandig mojlig hoger "
    ),
    "tr": (
        "klasor klasoru surucu suruculer yukle yukleyici degistir degil baslat basarili goster "
        "duzenle tum tumu bolum kayit diger bulunamadi tasinabilir calistir kaldir yonetici simdi "
        "surum guncelleme guncellemeler ozellik ozellikler islem lutfen secin secili "
    ),
}


def _folded_words(lang):
    return set(" ".join(FOLDED.get(lang, ())).split()) if isinstance(FOLDED.get(lang), tuple) \
        else set(FOLDED.get(lang, "").split())


# letters: characters that do not belong to the language.
POLISH_ONLY = "łńśźżŁŃŚŹŻ"
NON_LATIN = ("bg", "ru", "uk", "el", "zh", "ja", "ko")
LATIN_LANGS = [lang for lang in LANGS if lang not in NON_LATIN]
FORBIDDEN = {
    "uk": "ыэъёЫЭЪЁ",          # Russian letters
    "bg": "ыэёіїєґЫЭЁІЇЄҐ",    # Russian/Ukrainian letters
    "ru": "іїєґІЇЄҐ",          # Ukrainian letters
    "ro": "şţŞŢãÃ",            # cedilla instead of comma (ș ț); ã instead of ă
    "hu": "õûÕÛ",              # tilde/circumflex instead of double acute (ő ű)
}
CYRILLIC_RE = re.compile("[Ѐ-ӿ]")
GREEK_RE = re.compile("[Ͱ-Ͽἀ-῿]")

# greek_tonos: Greek words of 5+ letters without a stress mark (all-caps
# words are written without accents).
GREEK_WORD_RE = re.compile("[Ͱ-Ͽ]{5,}")
GREEK_ACCENT_RE = re.compile("[άέήίόύώΆΈΉΊΌΎΏΐΰ]")
GREEK_NO_TONOS_OK = {"στους", "στων", "στις", "στην", "στον", "στης"}  # one syllable


def share_problems(data):
    out = {}
    for lang, minimum in SHARE_MIN.items():
        letters = set(SPECIAL[lang])
        long_texts = [v for v in data[lang].values() if len(words(v)) >= SHARE_MIN_WORDS]
        if not long_texts:
            continue
        with_special = sum(1 for v in long_texts if letters & set(v.lower()))
        share = with_special / len(long_texts)
        if share < minimum:
            out[lang] = [f"only {with_special}/{len(long_texts)} longer texts ({share:.0%}) contain any of "
                         f"{''.join(sorted(letters))} (minimum {minimum:.0%}): folded to ASCII?"]
    return out


def folded_problems(data):
    out = {}
    for lang in FOLDED:
        bad = _folded_words(lang)
        msgs = [f"{key}: folded word {w!r} in {value!r}"
                for key, value in data[lang].items() for w in words(value) if w.lower() in bad]
        if msgs:
            out[lang] = msgs
    return out


def letter_problems(data):
    out = {}
    for lang in LANGS:
        msgs = []
        for key, value in data[lang].items():
            text = _NOISE_RE.sub(" ", value)
            bad = sorted(set(c for c in text if c in FORBIDDEN.get(lang, "")))
            if bad:
                msgs.append(f"{key}: letters {''.join(bad)} do not belong to {lang} in {value!r}")
            if lang in LATIN_LANGS and lang != "pl":
                bad = sorted(set(c for c in text if c in POLISH_ONLY))
                if bad:
                    msgs.append(f"{key}: Polish letters {''.join(bad)} in {value!r}")
            if lang in LATIN_LANGS and (CYRILLIC_RE.search(text) or GREEK_RE.search(text)):
                msgs.append(f"{key}: Cyrillic/Greek letters in {value!r}")
            if lang in ("bg", "ru", "uk") and GREEK_RE.search(text):
                msgs.append(f"{key}: Greek letters in {value!r}")
            if lang == "el" and CYRILLIC_RE.search(text):
                msgs.append(f"{key}: Cyrillic letters in {value!r}")
        if msgs:
            out[lang] = msgs
    return out


def greek_problems(data):
    msgs = []
    for key, value in data["el"].items():
        for w in GREEK_WORD_RE.findall(_NOISE_RE.sub(" ", value)):
            if w.isupper() or w in GREEK_NO_TONOS_OK or GREEK_ACCENT_RE.search(w):
                continue
            msgs.append(f"{key}: {w!r} without a stress mark in {value!r}")
    return {"el": msgs} if msgs else {}


CHECKS = {"share": share_problems, "folded": folded_problems,
          "letters": letter_problems, "greek_tonos": greek_problems}


class DiacriticsTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.data = corpus()

    def test_corpus_has_every_language(self):
        for lang in LANGS:
            self.assertGreater(len(self.data[lang]), 500, lang)

    def test_lists_are_consistent(self):
        for lang in list(FOLDED) + list(SHARE_MIN):
            self.assertIn(lang, SPECIAL)
        for lang in FOLDED:
            for w in _folded_words(lang):
                self.assertTrue(w.isascii() and w.islower(), (lang, w))

    def test_checks_catch_folded_text(self):
        # A language folded to ASCII must fail every Latin-script check.
        sample = {lang: dict(texts) for lang, texts in self.data.items()}
        from unicodedata import category, normalize
        sample["pl"] = {k: "".join(c for c in normalize("NFD", v.replace("ł", "l"))
                                    if category(c) != "Mn")
                        for k, v in sample["pl"].items()}
        self.assertIn("pl", share_problems(sample))
        self.assertIn("pl", folded_problems(sample))

    def _run(self, name):
        problems = CHECKS[name](self.data)
        for lang in LANGS:
            with self.subTest(lang=lang):
                msgs = problems.get(lang, [])
                self.assertEqual(msgs, [], f"{lang}: {len(msgs)} problem(s)\n  " + "\n  ".join(msgs[:40]))

    def test_share_of_texts_with_native_letters(self):
        self._run("share")

    def test_no_typical_folded_words(self):
        self._run("folded")

    def test_letters_belong_to_the_language(self):
        self._run("letters")

    def test_greek_words_carry_stress_marks(self):
        self._run("greek_tonos")


if __name__ == "__main__":
    if sys.argv[1:] == ["report"]:
        sys.stdout.reconfigure(encoding="utf-8")
        data = corpus()
        for name, check in CHECKS.items():
            for lang, msgs in check(data).items():
                for m in msgs:
                    print(f"[{name}] {lang} {m}")
        for lang, letters in SPECIAL.items():
            long_texts = [v for v in data[lang].values() if len(words(v)) >= SHARE_MIN_WORDS]
            n = sum(1 for v in long_texts if set(letters) & set(v.lower()))
            print(f"[share-stat] {lang} {n}/{len(long_texts)} = {n / max(1, len(long_texts)):.0%}")
    else:
        unittest.main()
