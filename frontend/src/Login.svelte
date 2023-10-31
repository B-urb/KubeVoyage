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

      const data = await response.json();

      if (response.ok) {
        message = "Login successful!";
        if (data.redirect) {
          isRedirecting = true;
          const authResponse = await fetch('/api/authenticate', {
            method: 'GET',
            credentials: 'include'
          });

          if (authResponse.status === 200) {
            setTimeout(() => {
              window.location.href = "/api/redirect";
            }, 2000); //FIXME Redirect
          } else if (authResponse.status === 401) {
            await fetch('/api/request', {
              method: 'POST',
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json'
              },
              body: JSON.stringify({ email })
            });
            message = "Access request sent. Please wait for approval.";
          } else {
            message = "Unexpected error occurred. Please try again later.";
          }
        }
        else {
          navigate("/")
        }
      } else {
        message = data.error || "Login failed!";
      }
    } catch (error) {
      message = "An error occurred: " + error.message;
    }
  }

  function handleSubmit(event) {
    event.preventDefault();
    login();
  }
</script>

<div class="container mt-5">
  <div class="row justify-content-center">
    <div class="col-md-4">
      {#if !isRedirecting}
        <h2>Login</h2>
        <form on:submit={handleSubmit}>
          <div class="mb-3">
            <label for="email" class="form-label">Email address</label>
            <input type="email" class="form-control" id="email" bind:value={email}>
          </div>
          <div class="mb-3">
            <label for="password" class="form-label">Password</label>
            <input type="password" class="form-control" id="password" bind:value={password}>
          </div>
          <button type="submit" class="btn btn-primary">Login</button>
        </form>
      {:else}
        <div class="text-center">
          <div class="spinner-border" role="status">
          </div>
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
