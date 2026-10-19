// Past stocks: bought and sold in full, with realized profit/loss. Reloads after each trade refresh.
import { api, usd, num, showMessage, hideMessage, cell, pickable, markShown } from './common.js';

const $ = (id) => document.getElementById(id);
const body = $('closed-body');
const msg = $('closed-msg');
let shown = ''; // symbol on the chart

async function loadClosed() {
  // Keep a buy result: the buy itself starts this reload.
  if (msg.classList.contains('alert-danger')) hideMessage(msg);
  try {
    const { closed } = await api('/api/closed');
    body.replaceChildren();
    let total = 0;
    for (const c of closed) {
      const pl = Number(c.realizedPl);
      total += pl;
      const tr = document.createElement('tr');
      pickable(tr, c.symbol);
      const symbolCell = cell(c.symbol, { ltr: true });
      symbolCell.prepend(buyButton(c), ' ');
      tr.append(
        symbolCell,
        cell(num(c.qty), { ltr: true }),
        cell(usd(c.avgBuy), { ltr: true }),
        cell(usd(c.avgSell), { ltr: true }),
        cell(usd(c.realizedPl), { ltr: true, cls: pl >= 0 ? 'text-up' : 'text-down' }),
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

// buyButton buys the same quantity again with a market (day) order.
function buyButton(c) {
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'btn btn-sm btn-outline-success py-0';
  btn.textContent = 'קנה';
  btn.setAttribute('aria-label', `קנה ${num(c.qty)} ${c.symbol} במחיר שוק`);
  btn.addEventListener('click', async (e) => {
    e.stopPropagation(); // buy only, do not open the symbol
    if (!confirm(`לקנות ${num(c.qty)} ${c.symbol} במחיר שוק?`)) return;
    hideMessage(msg);
    btn.disabled = true;
    try {
      const o = await api('/api/orders', {
        method: 'POST',
        body: JSON.stringify({ symbol: c.symbol, side: 'buy', type: 'market', qty: c.qty }),
      });
      showMessage(msg, `נשלחה קנייה במחיר שוק: ${num(o.qty)} ${o.symbol}`, 'success');
      document.dispatchEvent(new Event('orders-changed')); // trade.js reloads, then this table
    } catch (err) {
      showMessage(msg, err.message);
      btn.disabled = false;
    }
  });
  return btn;
}

document.addEventListener('symbol-loaded', (e) => {
  shown = e.detail;
  markShown(body, shown);
});
document.addEventListener('trades-changed', loadClosed); // sent by trade.js after each refresh
