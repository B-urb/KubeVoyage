<script>
 import { onMount } from 'svelte';

  let requests = [];

  onMount(async () => {
    try {
      const response = await fetch('http://localhost:8080/api/requests');
      if (response.ok) {
        requests = await response.json();
      } else {
        console.error('Failed to fetch requests:', response.statusText);
      }
    } catch (error) {
      console.error('Error fetching requests:', error);
    }
  });
  function acceptRequest(id) {
    // Handle accept logic here
  }

  function denyRequest(id) {
    // Handle deny logic here
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
    {#each requests as request, index (index)}
        <tr>
          <td>{request.user}</td>
          <td>{request.site}</td>
          <td>
            <button class="btn btn-success btn-sm" on:click={() => acceptRequest(request.id)}>Accept</button>
            <button class="btn btn-danger btn-sm" on:click={() => denyRequest(request.id)}>Deny</button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>