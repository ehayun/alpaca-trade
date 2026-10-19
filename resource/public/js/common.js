// Shared helpers: API calls and Hebrew/US formatting.

// api calls a JSON endpoint wrapped by GoFrame's {code, message, data} envelope.
export async function api(path, options = {}) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (res.status === 401) {
    location.href = '/login'; // the 60-minute login cookie expired
    throw new Error('נדרשת התחברות');
  }
  let body;
  try {
    body = await res.json();
  } catch {
    throw new Error('שגיאת שרת, נסו שוב');
  }
  if (body.code !== 0) throw new Error(body.message || 'שגיאה לא צפויה');
  return body.data;
}

const usdFmt = new Intl.NumberFormat('he-IL', { style: 'currency', currency: 'USD' });
const numFmt = new Intl.NumberFormat('he-IL', { maximumFractionDigits: 4 });
const dateTimeFmt = new Intl.DateTimeFormat('he-IL', {
  timeZone: 'Asia/Jerusalem',
  day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit',
});

export const usd = (v) => (v === null || v === undefined || v === '' ? '-' : usdFmt.format(Number(v)));
export const num = (v) => (v === null || v === undefined || v === '' ? '-' : numFmt.format(Number(v)));
export const dateTime = (s) => dateTimeFmt.format(new Date(s));

// showMessage puts text into a Bootstrap alert element and reveals it.
export function showMessage(el, text, kind = 'danger') {
  el.className = `alert alert-${kind} alert-dismissible`;
  el.textContent = text;
  // Bootstrap's JS is not loaded, so the close button gets its own handler.
  const close = document.createElement('button');
  close.type = 'button';
  close.className = 'btn-close';
  close.setAttribute('aria-label', 'סגירה');
  close.addEventListener('click', () => hideMessage(el));
  el.append(close);
  clearTimeout(hideTimers.get(el)); // a new message gets the full 10 seconds
  hideTimers.set(el, setTimeout(() => hideMessage(el), 10_000));
}
const hideTimers = new WeakMap(); // message element -> auto-hide timer

export function hideMessage(el) {
  el.className = 'alert d-none';
  el.textContent = '';
}

// showTab shows the panel of a tab and hides the other panels of its tab list.
function showTab(tab) {
  tab.closest('[role=tablist]').querySelectorAll('[role=tab]').forEach((t) => {
    const on = t === tab;
    t.classList.toggle('active', on);
    t.setAttribute('aria-selected', on);
    document.getElementById(t.getAttribute('aria-controls')).hidden = !on;
  });
}
// One listener for all tab lists on the page (side panel tabs and chart/info tabs).
document.addEventListener('click', (e) => {
  const tab = e.target.closest('[role=tab]');
  if (tab) showTab(tab);
});

// openSymbol loads a symbol into the chart (and so the info tabs and the order form).
export function openSymbol(symbol) {
  const form = document.getElementById('bars-form');
  form.symbol.value = symbol;
  form.requestSubmit();
  // Bring the chart into view when the row that was clicked is far from it.
  const top = form.getBoundingClientRect().top;
  if (top < 0 || top > innerHeight) form.scrollIntoView({ behavior: 'smooth' });
}

// pickable makes a table row open its symbol on click, Enter or Space.
export function pickable(tr, symbol) {
  tr.dataset.symbol = symbol;
  tr.tabIndex = 0;
  tr.classList.add('row-pick');
  tr.addEventListener('click', () => openSymbol(symbol));
  tr.addEventListener('keydown', (e) => {
    if (e.target === tr && (e.key === 'Enter' || e.key === ' ')) {
      e.preventDefault();
      openSymbol(symbol);
    }
  });
}

// Up/Down arrows move between the rows of a list; the chart follows when the key stops (no request per row).
let arrowTimer;
document.addEventListener('keydown', (e) => {
  const tr = e.target;
  if ((e.key !== 'ArrowDown' && e.key !== 'ArrowUp') || !tr.classList?.contains('row-pick')) return;
  let next = tr;
  do next = e.key === 'ArrowDown' ? next.nextElementSibling : next.previousElementSibling;
  while (next && next.hidden); // skip rows the search filter hides
  e.preventDefault(); // no page scroll
  if (!next) return;
  next.focus();
  clearTimeout(arrowTimer);
  arrowTimer = setTimeout(() => openSymbol(next.dataset.symbol), 250);
});

// markShown highlights the rows of the symbol that is on the chart.
export function markShown(body, symbol) {
  for (const tr of body.children) {
    const on = tr.dataset.symbol === symbol;
    tr.classList.toggle('table-primary', on);
    if (on) tr.setAttribute('aria-current', 'true');
    else tr.removeAttribute('aria-current');
  }
}

// cell builds a <td>; numbers and symbols are isolated so they don't flip inside RTL text.
export function cell(text, { ltr = false, cls = '' } = {}) {
  const td = document.createElement('td');
  if (cls) td.className = cls;
  if (ltr) {
    const b = document.createElement('bdi');
    b.textContent = text;
    td.append(b);
  } else {
    td.textContent = text;
  }
  return td;
}
