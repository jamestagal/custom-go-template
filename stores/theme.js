export default {
  mode: 'light',
  colors: {
    light: {
      background: '#ffffff',
      text: '#1a202c',
      primary: '#3182ce',
      secondary: '#718096'
    },
    dark: {
      background: '#1a202c',
      text: '#f7fafc',
      primary: '#63b3ed',
      secondary: '#a0aec0'
    }
  },
  getCurrentColors() {
    return this.colors[this.mode];
  },
  setLight() {
    this.mode = 'light';
    localStorage.setItem('theme', 'light');
  },
  setDark() {
    this.mode = 'dark';
    localStorage.setItem('theme', 'dark');
  },
  toggle() {
    this.mode = this.mode === 'light' ? 'dark' : 'light';
    localStorage.setItem('theme', this.mode);
  },
  init() {
    const saved = localStorage.getItem('theme');
    if (saved) {
      this.mode = saved;
    } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      this.mode = 'dark';
    }
  }
}
