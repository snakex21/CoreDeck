// Language detection and DOM translation runtime.
// --- Language detection & application ---

function detectLanguage() {
  // 1. Check localStorage for manual override
  const saved = localStorage.getItem('launcher-lang');
  if (saved && I18N[saved]) return saved;

  // 2. Auto-detect from browser/OS
  const browserLangs = navigator.languages || [navigator.language || navigator.userLanguage || 'pl'];
  for (const lang of browserLangs) {
    const code = lang.split('-')[0].toLowerCase();
    if (I18N[code]) return code;
  }
  return 'pl'; // fallback
}

let currentLang = detectLanguage();

function t(key) {
  return (I18N[currentLang] && I18N[currentLang][key]) || (I18N['pl'] && I18N['pl'][key]) || key;
}

function setLanguage(lang) {
  if (!I18N[lang]) return;
  currentLang = lang;
  localStorage.setItem('launcher-lang', lang);
  applyTranslations();
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
}

// Expose globally
window.I18N = I18N;
window.t = t;
window.setLanguage = setLanguage;
window.applyTranslations = applyTranslations;
window.getCurrentLang = () => currentLang;
