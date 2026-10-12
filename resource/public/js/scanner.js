// EMA scanner side panel: choose the spread limit and whether the price must be close too, start a scan,
// show progress, list the matches with live prices; a click loads the chart and info.
// On open, a saved scan (if any) is offered instead of a new scan.
import { api, showMessage, hideMessage, num, usd, dateTime, pickable, markShown } from './common.js';

const $ = (id) => document.getElementById(id);
const btn = $('scan-btn');
const statusEl = $('scan-status');
const msg = $('scan-msg');
const list = $('scan-list');
const saved = $('scan-saved');
const pctInput = $('scan-pct');
const priceClose = $('scan-price-close');


const POLL_MS = 2000;
let shown = ''; // symbol on the chart, highlighted in the list

// The settings a scan ran with, e.g. "עד 2% כולל המחיר".
function settingsText(s) {
  return `עד ${num(s.maxSpreadPct)}%${s.priceClose ? ' כולל המחיר' : ''}`;
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

// The spread a scan used: with the close as a fourth value when the price rule was on.
function row(m, withPrice) {
  const tr = document.createElement('tr');
  tr.title = `EMA 20: ${m.ema20.toFixed(2)} · EMA 100: ${m.ema100.toFixed(2)} · EMA 200: ${m.ema200.toFixed(2)}`
    + ` · סגירה: ${m.close.toFixed(2)}`;
  tr.append(
    td(m.symbol, 'fw-medium'),
    td(usd(m.close), 'text-end scan-price'), // close at scan time until the live price arrives
    td(`${(withPrice ? m.priceSpreadPct : m.spreadPct).toFixed(2)}%`, 'text-end text-body-secondary'),
  );
  pickable(tr, m.symbol);
  return tr;
}

function highlight() {
  markShown(list, shown);
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
