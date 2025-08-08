<script lang="ts">
  import type { QueryOutput } from "@/store/store";
  let { output = $bindable() }: { output: QueryOutput | null } = $props();
</script>

<div>
  {#if output}
    <table>
      <thead>
        <tr>
          {#each output.columns as column}
            <th>{column}</th>
          {/each}
        </tr>
      </thead>
      <tbody>
        {#each output.data as row}
          <tr>
            {#each output.columns as column}
              <td title={row[column]}>{row[column]}</td>
            {/each}
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: small;
  }
  th,
  td {
    border: 1px solid #ddd;
    padding: 8px;
    text-align: left;
    max-width: 50px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  th {
    background-color: #f2f2f2;
  }
  tr:nth-child(even) {
    background-color: #f9f9f9;
  }
</style>
