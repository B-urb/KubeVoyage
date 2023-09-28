<script>
  import 'bootstrap-icons/font/bootstrap-icons.css';

  let email = '';
  let password = '';

  function login() {
    // Handle login logic here
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

    <style>
      .sso-login {
        text-align: center;
      }
      .sso-login .btn {
        margin: 0 5px;
      }
    </style>
  </div>
</div>