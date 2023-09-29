<script>

  import {navigate} from "svelte-routing";

  let email = '';
  let password = '';
  let message = '';

  async function login() {
    try {
      const response = await fetch('http://localhost:8080/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email, password })
      });

      const data = await response.json();

      if (response.ok) {
        message = "Login successful!";
        navigate("/")
        // Optionally, set a token, redirect the user, or perform other actions
        // For example: localStorage.setItem('token', data.token);
      } else {
        message = data.error || "Login failed!";
      }
    } catch (error) {
      message = "An error occurred: " + error.message;
    }
  }
</script>

<div class="container mt-5">
  <div class="row justify-content-center">
    <div class="col-md-4">
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