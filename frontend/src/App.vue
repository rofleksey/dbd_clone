<template>
  <div id="app-root">
    <router-view />
  </div>
</template>

<script setup lang="ts">
</script>

<style>
#app-root {
  width: 100vw;
  height: 100vh;
  background: #0a0a0a;
}

/* Global styles */
:root {
  --primary: #c62828;
  --primary-dark: #8e0000;
  --primary-light: #ff5f52;
  --bg-dark: #0a0a0a;
  --bg-card: #1a1a1a;
  --bg-hover: #2a2a2a;
  --border: #333;
  --text: #e0e0e0;
  --text-dim: #888;
  --success: #2e7d32;
  --warning: #f57f17;
  --danger: #c62828;
}

button {
  cursor: pointer;
  border: none;
  padding: 10px 24px;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
}

button:hover {
  transform: translateY(-1px);
}

button:active {
  transform: translateY(0);
}

.btn-primary {
  background: var(--primary);
  color: white;
}

.btn-primary:hover {
  background: var(--primary-light);
}

.btn-secondary {
  background: var(--bg-hover);
  color: var(--text);
  border: 1px solid var(--border);
}

.btn-secondary:hover {
  background: #3a3a3a;
}

.btn-success {
  background: var(--success);
  color: white;
}

.btn-danger {
  background: var(--danger);
  color: white;
}

input {
  background: var(--bg-card);
  border: 1px solid var(--border);
  color: var(--text);
  padding: 10px 14px;
  border-radius: 4px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

input:focus {
  border-color: var(--primary);
}

a {
  color: var(--primary-light);
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}
</style>
