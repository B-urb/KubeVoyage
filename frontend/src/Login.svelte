<script>
  import { navigate } from "svelte-routing";

  let email = '';
  let password = '';
  let message = '';
  let isRedirecting = false;

  async function login() {
    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        credentials: "include",
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email, password })
      });

      if (response.ok) {
        message = "Login successful!";
        isRedirecting = true;
        setTimeout(() => {
          window.location.href = "/api/redirect";
        }, 2000);
      } else {
        message = response.error || "Login failed!";
      }
    } catch (error) {
      message = "An error occurred: " + error.message;
    }
  }
</script>

<div class="container mt-5">
  <div class="row justify-content-center">
    <div class="col-md-4">
      {#if !isRedirecting}
        <h2>Login</h2>
        <form>
          <div class="mb-3">
            <label for="email" class="form-label">Email address</label>
            <input type="email" class="form-control" id="email" bind:value={email}>
          </div>
          <div class="mb-3">
            <label for="password" class="form-label">Password</label>
            <input type="password" class="form-control" id="password" bind:value={password}>
          </div>
          <button type="button" class="btn btn-primary" on:click={login}>Login</button>
        </form>
      {:else}
        <div class="text-center">
          <span class="sr-only">Loading...</span>
          <div class="spinner-border" role="status"/>
          <p class="mt-3">Redirecting, please wait...</p>
        </div>
      {/if}
    </div>
    <div class="sso-login mt-4">
      <p>Or login with:</p>
      <a href="/auth/google" class="btn btn-light">
        <i class="bi bi-google"></i> Google
      </a>
      <a href="/auth/github" class="btn btn-light">
        <i class="bi bi-github"></i> GitHub
      </a>
      <a href="/auth/microsoft" class="btn btn-light">
        <i class="bi bi-windows"></i> Microsoft
      </a>
    </div>
  </div>
</div>

<style>
  .sso-login {
    text-align: center;
  }
  .sso-login .btn {
    margin: 0 5px;
  }
</style>
