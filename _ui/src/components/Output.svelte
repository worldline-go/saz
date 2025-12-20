<script lang="ts">
  import {
    ArrowLeft,
    ArrowRight,
    Download,
    ExternalLink,
    X,
  } from "@lucide/svelte";
  import Rows from "./Rows.svelte";
  import { storeOutput } from "@/store/store";
  import { exportToCSV, exportToJSON, outputToData } from "@/helper/csv";
  import { tableHTML } from "@/helper/table";

  let offset = $state(0);
  let limit = $state(10);

  let downloadType = $state("csv");
  let downloadTypes = ["csv", "json"];

  const nextPage = () => {
    if ($storeOutput && $storeOutput.rows) {
      if (+limit + offset < $storeOutput?.rows?.length) {
        offset = Math.min(+limit + offset, $storeOutput?.rows?.length ?? 0);
      }
    }
  };

  const prevPage = () => {
    offset = Math.max(-limit + offset, 0);
  };

  const download = () => {
    let date = new Date();
    let dateExt = `${date.getFullYear()}${("0" + (date.getMonth() + 1)).slice(-2)}${("0" + date.getDate()).slice(-2)}`;

    switch (downloadType) {
      case "csv":
        exportToCSV($storeOutput, `output_${dateExt}.csv`);
        break;
      case "json":
        exportToJSON(outputToData($storeOutput), `output_${dateExt}.json`);
        break;
    }
  };

  const openInNewTab = () => {
    let htmlText = tableHTML($storeOutput);

    // Create a blob of the data
    const blob = new Blob([htmlText], { type: "text/html" });
    const url = URL.createObjectURL(blob);
    window.open(url, "_blank")?.focus();
  };

  $effect(() => {
    if ($storeOutput) {
      offset = 0;
    }
  });
</script>

<div class="grid">
  <div class="flex items-center bg-gray-300 justify-between">
    <div class="px-2 flex items-center">
      <span class="text-sm font-semibold">Query Output</span>
      <select
        class="select ml-2 mr-1 border-none rounded-none bg-gray-300 hover:cursor-pointer hover:bg-gray-100 w-28 h-6"
        bind:value={downloadType}
      >
        {#each downloadTypes as type}
          <option value={type}>{type}</option>
        {/each}
      </select>
      <button
        class="text-xs text-gray-500 hover:bg-yellow-200 hover:cursor-pointer px-1"
        onclick={download}
        title="Download Output"
      >
        <Download />
      </button>
      <button
        class="text-xs text-gray-500 hover:bg-yellow-200 hover:cursor-pointer px-1"
        onclick={openInNewTab}
        title="Open in New Temp Tab"
      >
        <ExternalLink />
      </button>
    </div>
    <div class="flex items-center">
      {#if $storeOutput?.rows_affected}
        <span class="text-xs text-gray-600 px-2">
          Rows Affected: {$storeOutput.rows_affected}
        </span>
      {/if}
      {#if $storeOutput?.duration}
        <span class="text-xs text-gray-600 px-2">
          Duration: {$storeOutput.duration}
        </span>
      {/if}
      {#if $storeOutput}
        <span class="text-xs text-gray-600 px-2">
          Offset: {offset}, Limit:
          <input
            type="text"
            size={`${limit}`.length || 1}
            bind:value={limit}
          />, Total: {$storeOutput?.rows?.length ?? 0}
        </span>
      {/if}
      <button
        class=" text-gray-500 hover:bg-yellow-200 hover:cursor-pointer"
        onclick={prevPage}
      >
        <ArrowLeft />
      </button>
      <button
        class=" text-gray-500 hover:bg-yellow-200 hover:cursor-pointer"
        onclick={nextPage}
      >
        <ArrowRight />
      </button>
      <button
        class="text-gray-500 hover:bg-red-500 hover:text-white px-2 hover:cursor-pointer"
        onclick={() => {
          storeOutput.set(null);
        }}
      >
        <X />
      </button>
    </div>
  </div>
  <Rows output={$storeOutput} {offset} {limit} />
</div>
