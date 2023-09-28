<script>
  import { Router, Route, Link } from 'svelte-routing';
  import routes from './routes.js';
  import 'bootstrap/dist/css/bootstrap.min.css';
  import 'bootstrap-icons/font/bootstrap-icons.css';

</script>

<Router>
  <nav class="navbar navbar-expand-lg navbar-light bg-light">
    <a class="navbar-brand" href="#">KubeVoyage</a>
    <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
      <span class="navbar-toggler-icon"></span>
    </button>
    <div class="collapse navbar-collapse" id="navbarNav">
      <ul class="navbar-nav">
        <li class="nav-item active">
          <Link class="nav-link" to="/">Login</Link>
        </li>
        <li class="nav-item active">
          <Link class="nav-link" to="/register">Register</Link>
        </li>
        <li class="nav-item">
          <Link class="nav-link" to="/requests">Requests</Link>
        </li>
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