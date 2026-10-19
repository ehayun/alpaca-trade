// Past stocks: bought and sold in full, with realized profit/loss. Reloads after each trade refresh.
import { api, usd, num, showMessage, hideMessage, cell, pickable, markShown } from './common.js';

const $ = (id) => document.getElementById(id);
const body = $('closed-body');
const msg = $('closed-msg');
let shown = ''; // symbol on the chart

async function loadClosed() {
  hideMessage(msg);
  try {
    const { closed } = await api('/api/closed');
    body.replaceChildren();
    let total = 0;
    for (const c of closed) {
      const pl = Number(c.realizedPl);
      total += pl;
      const tr = document.createElement('tr');
      pickable(tr, c.symbol);
      tr.append(
        cell(c.symbol, { ltr: true }),
        cell(num(c.qty), { ltr: true }),
        cell(usd(c.avgBuy), { ltr: true }),
        cell(usd(c.avgSell), { ltr: true }),
        cell(usd(c.realizedPl), { ltr: true, cls: pl >= 0 ? 'text-up' : 'text-down' }),
        cell(c.day.split('-').reverse().join('/'), { ltr: true }), // DD/MM/YYYY
      );
      body.append(tr);
    }
    markShown(body, shown);
    const sum = $('closed-total');
    sum.textContent = usd(total);
    sum.className = total >= 0 ? 'text-up' : 'text-down';
    $('closed-foot').classList.toggle('d-none', closed.length === 0);
    $('closed-empty').classList.toggle('d-none', closed.length > 0);
  } catch (err) {
    showMessage(msg, err.message);
  }
}

document.addEventListener('symbol-loaded', (e) => {
  shown = e.detail;
  markShown(body, shown);
});
document.addEventListener('trades-changed', loadClosed); // sent by trade.js after each refresh
