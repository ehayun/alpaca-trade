// EMA scanner side panel: choose the spread limit and an optional stochastic rule, start a scan,
// show progress, list the matches with live prices; a click loads the chart and info.
// On open, a saved scan (if any) is offered instead of a new scan.
import { api, showMessage, hideMessage, num, usd, dateTime } from './common.js';

const $ = (id) => document.getElementById(id);
const btn = $('scan-btn');
const statusEl = $('scan-status');
const msg = $('scan-msg');
const list = $('scan-list');
const form = $('bars-form');
const saved = $('scan-saved');
const pctInput = $('scan-pct');
const stochOn = $('scan-stoch-on');
const stochRule = $('scan-stoch');
const priceClose = $('scan-price-close');

const STOCH_LABEL = { oversold: 'מכירת יתר', overbought: 'קניית יתר', bullcross: 'חציה עולה מתחת ל-20' };

const POLL_MS = 2000;
const stacked = matchMedia('(max-width: 991.98px)'); // Bootstrap lg breakpoint
let shown = ''; // symbol on the chart, highlighted in the list

// The settings a scan ran with, e.g. "עד 2% · סטוכסטי: מכירת יתר".
function settingsText(s) {
  const stoch = s.stochRule ? ` · סטוכסטי: ${STOCH_LABEL[s.stochRule] ?? s.stochRule}` : '';
  const price = s.priceClose ? ' כולל המחיר' : '';
  return `עד ${num(s.maxSpreadPct)}%${price}${stoch}`;
}

function statusText(s) {
  if (s.running) return `סורק… ${num(s.done)} מתוך ${num(s.total)} מניות (${settingsText(s)})`;
  if (!s.finishedAt) return 'עדיין לא נסרק.';
  return `נסרקו ${num(s.scanned)} מניות · ${dateTime(s.finishedAt)} · נמצאו ${num(s.matches.length)} (${settingsText(s)})`;
}

function td(text, cls = '') {
  const cell = document.createElement('td');
  cell.className = cls;
  const b = document.createElement('bdi');
  b.textContent = text;
  cell.append(b);
  return cell;
}

function openSymbol(symbol) {
  form.symbol.value = symbol;
  form.requestSubmit();
  // When the panel is stacked above the chart (narrow screens), bring the chart into view.
  if (stacked.matches) form.scrollIntoView({ behavior: 'smooth' });
}

// The spread a scan used: with the close as a fourth value when the price rule was on.
function row(m, withPrice) {
  const tr = document.createElement('tr');
  tr.dataset.symbol = m.symbol;
  tr.tabIndex = 0; // rows open with Enter or Space too
  tr.title = `EMA 20: ${m.ema20.toFixed(2)} · EMA 100: ${m.ema100.toFixed(2)} · EMA 200: ${m.ema200.toFixed(2)}`
    + ` · סגירה: ${m.close.toFixed(2)}`;
  tr.append(
    td(m.symbol, 'fw-medium'),
    td(usd(m.close), 'text-end scan-price'), // close at scan time until the live price arrives
    td(m.stochK === undefined ? '-' : m.stochK.toFixed(0), 'text-end'), // scans saved before the stochastic have no %K
    td(`${(withPrice ? m.priceSpreadPct : m.spreadPct).toFixed(2)}%`, 'text-end text-body-secondary'),
  );
  tr.addEventListener('click', () => openSymbol(m.symbol));
  tr.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      openSymbol(m.symbol);
    }
  });
  return tr;
}

function highlight() {
  for (const tr of list.children) {
    const on = tr.dataset.symbol === shown;
    tr.classList.toggle('table-primary', on);
    if (on) tr.setAttribute('aria-current', 'true');
    else tr.removeAttribute('aria-current');
  }
}

// Live prices for the listed symbols, in one request. On failure the scan-time close stays.
async function refreshPrices(matches) {
  if (!matches.length) return;
  try {
    const { prices } = await api(`/api/prices?symbols=${encodeURIComponent(matches.map((m) => m.symbol).join(','))}`);
    for (const tr of list.children) {
      const p = prices[tr.dataset.symbol];
      if (p !== undefined) tr.querySelector('.scan-price bdi').textContent = usd(p);
    }
  } catch (err) {
    console.warn('live prices:', err.message);
  }
}

function render(s) {
  btn.disabled = s.running;
  statusEl.textContent = statusText(s);
  if (s.failed) showMessage(msg, 'הסריקה נכשלה, נסו שוב');
  else hideMessage(msg);
  if (!s.running) {
    list.replaceChildren(...s.matches.map((m) => row(m, s.priceClose)));
    $('scan-table').hidden = s.matches.length === 0;
    highlight();
    refreshPrices(s.matches);
  }
}

// Poll until the scan finishes; the server keeps scanning even if the page is closed.
async function follow(s) {
  render(s);
  while (s.running) {
    await new Promise((r) => setTimeout(r, POLL_MS));
    try {
      s = await api('/api/scan');
    } catch (err) {
      showMessage(msg, err.message);
      btn.disabled = false;
      return;
    }
    render(s);
  }
}

async function startScan() {
  if (!pctInput.checkValidity()) {
    showMessage(msg, 'הפער חייב להיות בין 0.1% ל-20%');
    pctInput.focus();
    return;
  }
  saved.hidden = true;
  btn.disabled = true;
  const body = {
    maxSpreadPct: Number(pctInput.value),
    stoch: stochOn.checked ? stochRule.value : '',
    priceClose: priceClose.checked,
  };
  try {
    follow(await api('/api/scan', { method: 'POST', body: JSON.stringify(body) }));
  } catch (err) {
    showMessage(msg, err.message);
    btn.disabled = false;
  }
}

btn.addEventListener('click', startScan);
stochOn.addEventListener('change', () => {
  stochRule.disabled = !stochOn.checked;
});
$('scan-saved-rescan').addEventListener('click', startScan);

document.addEventListener('symbol-loaded', (e) => {
  shown = e.detail;
  highlight();
});

// On open: follow a running scan; otherwise offer the saved scan, if there is one.
async function init() {
  const s = await api('/api/scan');
  if (s.running) return follow(s);
  const last = await api('/api/scan/saved');
  if (!last) return render(s);
  $('scan-saved-text').textContent =
    `נמצאה סריקה שמורה מ-${dateTime(last.finishedAt)} (${num(last.matches.length)} מניות, ${settingsText(last)}).`;
  $('scan-saved-use').onclick = () => {
    saved.hidden = true;
    render(last);
  };
  statusEl.textContent = '';
  saved.hidden = false;
}

init().catch((err) => showMessage(msg, err.message));
