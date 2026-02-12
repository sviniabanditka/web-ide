<template>
  <router-view />
</template>

<style>
@import 'xterm/css/xterm.css';

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  width: 100%;
}

body {
  background: hsl(var(--background));
  color: hsl(var(--foreground));
}

input, textarea {
  font-family: inherit;
}

::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

::-webkit-scrollbar-track {
  background: hsl(var(--muted));
}

::-webkit-scrollbar-thumb {
  background: hsl(var(--border));
  border-radius: 5px;
}

::-webkit-scrollbar-thumb:hover {
  background: hsl(var(--muted-foreground));
}
</style>
