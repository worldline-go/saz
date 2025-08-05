<script lang="ts">
  let { rows = $bindable([]) }: { rows: Record<string, any>[] } = $props();
</script>

{#if rows.length > 0}
  <table>
    <thead>
      <tr>
        {#each Object.keys(rows[0]) as key}
          <th>{key}</th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#each rows as row}
        <tr>
          {#each Object.values(row) as value}
            <td>{value}</td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
{:else}
  <p>No rows to display.</p>
{/if}

<style>
  table {
    width: 100%;
    border-collapse: collapse;
  }
  th,
  td {
    border: 1px solid #ddd;
    padding: 8px;
    text-align: left;
  }
  th {
    background-color: #f2f2f2;
  }
  tr:nth-child(even) {
    background-color: #f9f9f9;
  }
</style>
