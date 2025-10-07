{
  isLoggedIn: false,
  user: null,
  login() {
    this.isLoggedIn = true;
    this.user = { name: 'Test User', email: 'test@example.com' };
  },
  logout() {
    this.isLoggedIn = false;
    this.user = null;
  }
}
