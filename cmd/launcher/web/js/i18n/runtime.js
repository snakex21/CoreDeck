// Language detection and DOM translation runtime.

// Interface languages, in the order shown in the picker (native names).
const LANGUAGES = [
  ['en', 'English'], ['pl', 'Polski'], ['bg', 'Български'], ['cs', 'Čeština'],
  ['da', 'Dansk'], ['de', 'Deutsch'], ['el', 'Ελληνικά'], ['es', 'Español'],
  ['et', 'Eesti'], ['fi', 'Suomi'], ['fr', 'Français'], ['hr', 'Hrvatski'],
  ['hu', 'Magyar'], ['it', 'Italiano'], ['lt', 'Lietuvių'], ['lv', 'Latviešu'],
  ['nb', 'Norsk bokmål'], ['nl', 'Nederlands'], ['pt', 'Português (Brasil)'],
  ['ro', 'Română'], ['ru', 'Русский'], ['sk', 'Slovenčina'], ['sl', 'Slovenščina'],
  ['sr-Latn', 'Srpski (latinica)'], ['sv', 'Svenska'], ['tr', 'Türkçe'],
  ['uk', 'Українська'], ['zh', '中文（简体）'], ['ja', '日本語'], ['ko', '한국어']
].filter(([code]) => I18N[code]);

// BCP 47 tags for Intl (plural rules, number formatting).
const LANGUAGE_TAGS = { pt: 'pt-BR', 'sr-Latn': 'sr-Latn', nb: 'nb', zh: 'zh-CN' };

// Maps a system/browser language tag (e.g. "pt-PT", "sr-Latn-RS", "nn-NO")
// to one of LANGUAGES, or "" when none fits.
function matchLanguage(tag) {
  const lower = String(tag || '').toLowerCase();
  const base = lower.split('-')[0];
  if (!base) return '';
  if (base === 'sr' || base === 'sh' || base === 'bs') return I18N['sr-Latn'] ? 'sr-Latn' : '';
  if (base === 'no' || base === 'nn' || base === 'nb') return I18N.nb ? 'nb' : '';
  if (base === 'cz') return 'cs';
  return I18N[base] ? base : '';
}

function savedLanguage() {
  try {
    const saved = localStorage.getItem('launcher-lang');
    return saved && I18N[saved] ? saved : '';
  } catch (_) {
    return '';
  }
}

function detectLanguage() {
  const saved = savedLanguage();
  if (saved) return saved;
  const browserLangs = navigator.languages || [navigator.language || navigator.userLanguage || 'en'];
  for (const lang of browserLangs) {
    const code = matchLanguage(lang);
    if (code) return code;
  }
  return 'en';
}

let currentLang = detectLanguage();

// t(key) -> text in the current language, then English, then Polish (the
// original), then the key. t(key, { name: value }) fills {name} placeholders.
function t(key, vars) {
  const text = (I18N[currentLang] && I18N[currentLang][key]) ||
    (I18N.en && I18N.en[key]) || (I18N.pl && I18N.pl[key]) || key;
  return vars ? fillPlaceholders(text, vars) : text;
}

function fillPlaceholders(text, vars) {
  return String(text).replace(/\{([a-z_]+)\}/g, (m, name) =>
    Object.prototype.hasOwnProperty.call(vars, name) ? String(vars[name]) : m);
}

// tn(base, n) -> plural form base_<one|few|many|other|...> chosen by the
// CLDR rules of the current language, with {n} filled in.
function tn(base, n, vars) {
  const lang = currentLang;
  let category = 'other';
  try {
    category = new Intl.PluralRules(LANGUAGE_TAGS[lang] || lang).select(n);
  } catch (_) { /* old engine: "other" */ }
  const table = I18N[lang] || {};
  const text = table[`${base}_${category}`] || table[`${base}_other`] ||
    (I18N.en && (I18N.en[`${base}_${new Intl.PluralRules('en').select(n)}`] || I18N.en[`${base}_other`])) || base;
  return fillPlaceholders(text, Object.assign({ n }, vars || {}));
}

function hasKey(key) {
  return Boolean((I18N[currentLang] && I18N[currentLang][key]) ||
    (I18N.en && I18N.en[key]) || (I18N.pl && I18N.pl[key]));
}

// Backend messages (internal/uimsg): Go sends a stable code - a key of this
// file - with params, next to the Polish text it keeps for logs. tMsg()
// translates the code (count_* codes are plural families, n in params) and
// falls back to the text for codes this build does not know.
function tMsg(code, params, fallback) {
  const vars = params || {};
  if (code && code.startsWith('count_') && hasKey(`${code}_other`)) {
    return tn(code, Number(vars.n) || 0, vars);
  }
  if (code && hasKey(code)) return t(code, vars);
  return fallback == null ? '' : String(fallback);
}

// tErr(response) -> translated error/message of an API response object.
function tErr(response) {
  if (!response) return '';
  return tMsg(response.code, response.params, response.error || response.message || '');
}

// "auto" or the language picked by hand.
function getLanguageChoice() {
  return savedLanguage() || 'auto';
}

function setLanguage(lang) {
  if (lang === 'auto') {
    try { localStorage.removeItem('launcher-lang'); } catch (_) { /* ignore */ }
    currentLang = detectLanguage();
  } else {
    if (!I18N[lang]) return;
    currentLang = lang;
    try { localStorage.setItem('launcher-lang', lang); } catch (_) { /* ignore */ }
  }
  applyTranslations();
}

// Fills every language picker (<select id="lang-select">) from LANGUAGES.
function populateLanguageSelects() {
  document.querySelectorAll('select#lang-select, select.lang-select').forEach(select => {
    select.textContent = '';
    select.add(new Option(t('lang_auto'), 'auto'));
    LANGUAGES.forEach(([code, name]) => select.add(new Option(name, code)));
    select.value = getLanguageChoice();
    select.setAttribute('aria-label', t('lang_label'));
    select.title = t('lang_label');
  });
}

function applyTranslations() {
  // Static HTML elements with data-i18n
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    const translated = t(key);
    if (el.tagName === 'INPUT' && el.hasAttribute('placeholder')) {
      el.placeholder = translated;
    } else if (el.tagName === 'TITLE') {
      document.title = translated;
    } else {
      // Preserve child elements (icons, etc.) if any
      const icon = el.querySelector('.mode-icon, .theme-icon, [class*="icon"]');
      if (icon) {
        el.textContent = '';
        el.appendChild(icon);
        el.append(' ' + translated);
      } else if (translated.includes('<')) {
        el.innerHTML = translated;
      } else {
        el.textContent = translated;
      }
    }
  });

  // data-i18n-placeholder for input/select
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    el.placeholder = t(el.getAttribute('data-i18n-placeholder'));
  });

  // data-i18n-title
  document.querySelectorAll('[data-i18n-title]').forEach(el => {
    el.title = t(el.getAttribute('data-i18n-title'));
  });

  // data-i18n-aria (aria-label)
  document.querySelectorAll('[data-i18n-aria]').forEach(el => {
    el.setAttribute('aria-label', t(el.getAttribute('data-i18n-aria')));
  });

  // data-i18n-alt
  document.querySelectorAll('[data-i18n-alt]').forEach(el => {
    el.alt = t(el.getAttribute('data-i18n-alt'));
  });

  // Apply to title
  document.title = t('title');
  document.documentElement.lang = LANGUAGE_TAGS[currentLang] || currentLang;
  const autoOption = document.querySelector('select#lang-select option[value="auto"]');
  if (autoOption) autoOption.textContent = t('lang_auto');
}

// Expose globally
window.I18N = I18N;
window.t = t;
window.setLanguage = setLanguage;
window.applyTranslations = applyTranslations;
window.getCurrentLang = () => currentLang;
window.tn = tn;
window.tMsg = tMsg;
window.tErr = tErr;
window.getLanguageChoice = getLanguageChoice;
window.LANGUAGES = LANGUAGES;

populateLanguageSelects();
