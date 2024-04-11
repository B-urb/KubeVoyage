<script>
  import { onMount } from 'svelte';
  let redirectURL = '';
  function getCookie(name) {
    let cookieArray = document.cookie.split(';');
    for (let i = 0; i < cookieArray.length; i++) {
      let cookiePair = cookieArray[i].split('=');
      if (name === cookiePair[0].trim()) {
        return decodeURIComponent(cookiePair[1]);
      }
    }
    return null;
  }

  onMount(() => {
    // Extract the redirect URL from the query parameters
    const urlParams = new URLSearchParams(window.location.search);
    redirectURL = urlParams.get('redirect');
  });

  async function requestAccess() {
    const response = await fetch('/api/request', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ redirect: redirectURL })
    });

    if (response.ok) {
      alert('Request submitted successfully!');
    } else {
      alert('Error submitting request.');
    }
  }
</script>

<div class="container mt-5">
  <h3>Request Access</h3>
  <p>You are trying to access: <strong>{getCookie('X-Auth-Site')}</strong></p>
  <button class="btn btn-primary" on:click={requestAccess}>Request Access</button>
</div>
