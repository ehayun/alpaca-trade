// Favorites tab: the symbols marked with a star in any list. A click on the star removes the row.
import { favoriteSymbols, symbolCell, pickable, markShown } from './common.js';

const body = document.getElementById('fav-body');
let shown = ''; // symbol on the chart

function render() {
  const symbols = favoriteSymbols();
  body.replaceChildren(...symbols.map((symbol) => {
    const tr = document.createElement('tr');
    pickable(tr, symbol);
    tr.append(symbolCell(symbol));
    return tr;
  }));
  markShown(body, shown);
  document.getElementById('fav-empty').classList.toggle('d-none', symbols.length > 0);
}

document.addEventListener('favorites-changed', render); // sent by common.js after the load and after each star click
document.addEventListener('symbol-loaded', (e) => {
  shown = e.detail;
  markShown(body, shown);
});
render(); // the favorites can arrive before this module runs
