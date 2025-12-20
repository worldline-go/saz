<script lang="ts">
  import type { QueryOutput } from "@/store/store";
  let {
    output = $bindable(),
    limit = $bindable(),
    offset = $bindable(),
  }: { output: QueryOutput | null; limit: number; offset: number } = $props();

  function handleCellClick(event: MouseEvent, value: string) {
    if (event.ctrlKey || event.metaKey) {
      navigator.clipboard.writeText(value);
      // Optional: Add visual feedback
      const target = event.currentTarget as HTMLElement;
      const originalBg = target.style.backgroundColor;
      target.style.backgroundColor = "#90EE90";
      setTimeout(() => {
        target.style.backgroundColor = originalBg;
      }, 200);
    }
  }
</script>

<div class="overflow-x-auto max-w-full">
  {#if output}
    <table>
      <thead>
        <tr>
          <th></th>
          {#each output.columns as column}
            <th>{column}</th>
          {/each}
        </tr>
      </thead>
      <tbody>
        {#if output.rows}
          {#each output?.rows?.slice(offset, +limit + offset) as row, index}
            <tr>
              <td>{index + offset + 1}</td>
              {#each row as value}
                <td
                  title={value}
                  onclick={(e) => handleCellClick(e, value)}
                  style="cursor: pointer;">{value}</td
                >
              {/each}
            </tr>
          {/each}
        {/if}
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
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  td:hover {
    overflow-x: auto;
    scrollbar-width: none;
    text-overflow: unset;
    -ms-overflow-style: none;
  }
  td:hover::-webkit-scrollbar {
    display: none;
  }
  th {
    background-color: #f2f2f2;
    position: sticky;
    top: 0;
    z-index: 1;
  }
  tr:nth-child(even) {
    background-color: #f9f9f9;
  }
  tr:hover {
    background-color: #fff085;
  }
  td:hover {
    background-color: #ffffcc;
  }
</style>
