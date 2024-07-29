<!-- Register.svelte -->

<script>
  import { navigate } from "svelte-routing";
  let email = '';
  let password = '';
  let confirmPassword = '';
  let message = '';  // To display any response or error messages

  async function register() {
    if (password !== confirmPassword) {
      message = "Passwords do not match!";
      return;
    }
    try {
      const response = await fetch('/api/register', {
        method: 'POST',
        credentials: "include",
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email, password})
      });

      const data = await response.json();

      if (response.ok) {
        message = "Registration successful!";
        navigate("/login")
      }
      else {
        message = data.error || "Registration failed!";
      }
    } catch (error) {
      message = "An error occurred: " + error.message;
    }
  }
</script>

<div class="container mt-5">
  <div class="row justify-content-center">
    <div class="col-md-6">
      <h2>Register</h2>
      <form on:submit|preventDefault={register}>
        <div class="form-group">
          <label for="email">Email</label>
          <input type="email" bind:value={email} class="form-control" id="email" placeholder="Enter email" required>
        </div>
        <div class="form-group">
          <label for="password">Password</label>
          <input type="password" bind:value={password} class="form-control" id="password" placeholder="Password" required>
        </div>
        <div class="form-group">
          <label for="confirmPassword">Confirm Password</label>
          <input type="password" bind:value={confirmPassword} class="form-control" id="confirmPassword" placeholder="Confirm Password" required>
        </div>
        <button type="submit" class="btn btn-primary">Register</button>
      </form>
    </div>
    {#if message}
      <div class="alert alert-danger" role="alert">
        {message}
      </div>
    {/if}
  </div>
</div>

<style>
  /* Add any additional styles if needed */
</style>
