<script>
  import { Router, Route, Link } from 'svelte-routing';
  import {onMount} from 'svelte';
  import { isAuthenticated } from './authStore.js';
  import routes from './routes.js';
  import 'bootstrap/dist/css/bootstrap.min.css';
  import 'bootstrap-icons/font/bootstrap-icons.css';



  onMount(async () => {
    isAuthenticated.checkAuth();
    await validateSession();
  });

  async function validateSession() {
    try {
      const response = await fetch('/api/validate-session', {
        credentials: 'include'  // This is important for sending cookies
      });
      if (!response.ok) {
        throw new Error('Session invalid');
      }
      isAuthenticated.setAuth(true);
    } catch (error) {
      console.error('Session validation failed:', error);
      isAuthenticated.setAuth(false);
    }
  }
  async function logout() {
    await fetch('/api/logout', { method: 'POST', credentials: 'include' });
    isAuthenticated.setAuth(false);
    // Handle logout (e.g., redirect to login page)
  }
</script>

<Router>
  <nav class="navbar navbar-expand-lg navbar-light bg-light">
    <a class="navbar-brand" href="#">KubeVoyage</a>
    <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
      <span class="navbar-toggler-icon"></span>
    </button>
    <div class="collapse navbar-collapse" id="navbarNav">
      <ul class="navbar-nav">
        {#if !$isAuthenticated}
        <li class="nav-item active">
          <Link class="nav-link" to="/login">Login</Link>
        </li>
        <li class="nav-item active">
          <Link class="nav-link" to="/register">Register</Link>
        </li>
          {/if}
        {#if $isAuthenticated}
        <li class="nav-item">
          <Link class="nav-link" to="/requests">Requests</Link>
        </li>
          <li>
          <button class="nav-link" on:click={logout}>Logout</button>
          </li>
          {/if}
        <!-- Add more links as needed -->
      </ul>
    </div>
  </nav>

  {#each Object.entries(routes) as [path, Component]}
    <Route path={path} let:params>
      <Component {...params} />
    </Route>
  {/each}
</Router>

<style>
  /* Add your styles here */
</style>