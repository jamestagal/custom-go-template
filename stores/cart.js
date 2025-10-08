{
  items: [],
  total: 0,
  get formattedTotal() {
    return this.total.toFixed(2);
  },
  addItem(item) {
    this.items.push(item);
    this.total += item.price;
  },
  removeItem(index) {
    const item = this.items[index];
    this.items.splice(index, 1);
    this.total -= item.price;
  },
  clear() {
    this.items = [];
    this.total = 0;
  }
}
