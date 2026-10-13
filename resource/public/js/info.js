// Chart/info tabs, and the info tab: company, today's price and news from /api/info.
import { api, showMessage, hideMessage, usd, num, dateTime } from './common.js';

const $ = (id) => document.getElementById(id);
const tabs = document.querySelectorAll('[role=tab]');
const status = $('info-status');
const msg = $('info-msg');
const body = $('info-body');

function showTab(tab) {
  tabs.forEach((t) => {
    const on = t === tab;
    t.classList.toggle('active', on);
    t.setAttribute('aria-selected', on);
    $(t.getAttribute('aria-controls')).hidden = !on;
  });
}
tabs.forEach((t) => t.addEventListener('click', () => showTab(t)));

// row appends a label/value pair, styled like the account card.
function row(dl, label, value, cls = '') {
  const dt = document.createElement('dt');
  dt.className = 'col-6 fw-normal text-body-secondary';
  dt.textContent = label;
  const dd = document.createElement('dd');
  dd.className = `col-6 text-end ${cls}`;
  const b = document.createElement('bdi');
  b.textContent = value;
  dd.append(b);
  dl.append(dt, dd);
}

const yesNo = (v) => (v ? 'כן' : 'לא');

// Arrow plus color, so the direction does not depend on color alone.
function change(q) {
  if (q.change === null) return ['-', ''];
  const pct = `${q.changePct.toFixed(2)}%`;
  if (q.change > 0) return [`▲ ${usd(q.change)} (+${pct})`, 'text-up'];
  if (q.change < 0) return [`▼ ${usd(-q.change)} (${pct})`, 'text-down'];
  return [`${usd(0)} (${pct})`, ''];
}

// Only http(s) links; anything else is shown as plain text.
function safeUrl(u) {
  try {
    const url = new URL(u);
    return url.protocol === 'https:' || url.protocol === 'http:' ? url : null;
  } catch {
    return null;
  }
}

function newsItem(n) {
  const url = safeUrl(n.url);
  const item = document.createElement(url ? 'a' : 'div');
  item.className = 'list-group-item' + (url ? ' list-group-item-action' : '');
  if (url) {
    item.href = url.href;
    item.target = '_blank';
    item.rel = 'noopener noreferrer';
  }
  const title = document.createElement('div');
  title.dir = 'auto';
  title.textContent = n.headline;
  const meta = document.createElement('small');
  meta.className = 'text-body-secondary';
  meta.textContent = dateTime(n.createdAt) + (url ? ` · ${url.hostname}` : '');
  item.append(title, meta);
  return item;
}

function render(info) {
  const a = info.asset;
  const asset = $('info-asset');
  asset.replaceChildren();
  row(asset, 'שם', a.name || '-');
  row(asset, 'בורסה', a.exchange || '-');
  row(asset, 'ניתן למסחר', yesNo(a.tradable));
  row(asset, 'ניתן למכירה בחסר', yesNo(a.shortable));
  row(asset, 'קנייה של חלק ממניה', yesNo(a.fractionable));

  const q = info.quote;
  const quote = $('info-quote');
  quote.replaceChildren();
  row(quote, 'מחיר אחרון', usd(q.price));
  row(quote, 'שינוי מסגירה קודמת', ...change(q));
  row(quote, 'סגירה קודמת', usd(q.prevClose));
  row(quote, 'פתיחה', usd(q.open));
  row(quote, 'גבוה', usd(q.high));
  row(quote, 'נמוך', usd(q.low));
  row(quote, 'מחזור', num(q.volume));

  const news = info.news || [];
  $('info-news').replaceChildren(...news.map(newsItem));
  $('info-news-empty').hidden = news.length > 0;
}

// Only the latest request may render, so a slow old answer can't replace a newer symbol.
let latest = 0;
let current = '';

const summaryBtn = $('summary-btn');
const summaryMsg = $('summary-msg');
const summaryBox = $('summary-box');

function resetSummary() {
  hideMessage(summaryMsg);
  summaryBox.hidden = true;
  summaryBox.textContent = '';
  summaryBtn.disabled = false;
}

// The summary is on demand: each one is a paid Claude call.
summaryBtn.addEventListener('click', async () => {
  const symbol = current;
  resetSummary();
  summaryBtn.disabled = true;
  summaryBox.hidden = false;
  summaryBox.textContent = 'מסכם את החדשות…';
  try {
    const data = await api(`/api/news-summary?symbol=${encodeURIComponent(symbol)}`);
    if (symbol !== current) return;
    summaryBox.textContent = data.summary;
  } catch (err) {
    if (symbol !== current) return;
    summaryBox.hidden = true;
    summaryBox.textContent = '';
    showMessage(summaryMsg, err.message);
  } finally {
    if (symbol === current) summaryBtn.disabled = false;
  }
});

async function load(symbol) {
  const id = ++latest;
  current = symbol;
  resetSummary();
  hideMessage(msg);
  body.hidden = true;
  status.hidden = false;
  status.textContent = 'טוען מידע…';
  try {
    const info = await api(`/api/info?symbol=${encodeURIComponent(symbol)}`);
    if (id !== latest) return;
    render(info);
    status.hidden = true;
    body.hidden = false;
  } catch (err) {
    if (id !== latest) return;
    status.hidden = true;
    showMessage(msg, err.message);
  }
}

document.addEventListener('symbol-loaded', (e) => load(e.detail));
