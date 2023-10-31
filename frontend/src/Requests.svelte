<script>
  import { onMount } from 'svelte';

  let requests = [];

  onMount(async () => {
    await fetchRequests();
  });

  async function fetchRequests() {
    try {
      const response = await fetch('/api/requests');
      if (response.ok) {
        requests = await response.json();
      } else {
        console.error('Failed to fetch requests:', response.statusText);
      }
    } catch (error) {
      console.error('Error fetching requests:', error);
    }
  }

  async function updateRequestState(userEmail, siteURL, newState) {
    try {
      const response = await fetch('/api/requests/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ userEmail, siteURL, newState })
      });

      if (response.ok) {
        await fetchRequests(); // Refetch the requests to update the UI
      } else {
        console.error('Failed to update request state:', response.statusText);
      }
    } catch (error) {
      console.error('Error updating request state:', error);
    }
  }

  function acceptRequest(request) {
    updateRequestState(request.user, request.site, 'authorized');
  }

  function denyRequest(request) {
    updateRequestState(request.user, request.site, 'declined');
  }
</script>

<div class="container mt-5">
  <h2>Access Requests</h2>
  <table class="table">
    <thead>
    <tr>
      <th>User</th>
      <th>Site</th>
      <th>Actions</th>
    </tr>
    </thead>
    <tbody>
    {#each requests as request (request.id)}
      <tr>
        <td>{request.user}</td>
        <td>{request.site}</td>
        <td>
          <button class="btn btn-success btn-sm" on:click={() => acceptRequest(request)}>Accept</button>
          <button class="btn btn-danger btn-sm" on:click={() => denyRequest(request)}>Deny</button>
        </td>
      </tr>
    {/each}
    </tbody>
  </table>
</div>
