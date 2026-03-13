<script lang="ts">
  import {
    ArrowLeft,
    ArrowRight,
    Download,
    ExternalLink,
    X,
  } from "@lucide/svelte";
  import Rows from "./Rows.svelte";
  import { storeOutput, storeNotebookOutput, type QueryOutput } from "@/store/store";
  import { exportToCSV, exportToJSON, outputToData } from "@/helper/csv";
  import { tableHTML } from "@/helper/table";

  let offset = $state(0);
  let limit = $state(10);
  let selectedCell = $state(0);

  let downloadType = $state("csv");
  let downloadTypes = ["csv", "json"];

  // Derive the active output from either the notebook cell selector or the single-query store
  let currentOutput = $derived.by(() => {
    if ($storeNotebookOutput && $storeNotebookOutput.length > 0) {
      const cell = $storeNotebookOutput[selectedCell] ?? $storeNotebookOutput[0];
      return {
        columns: cell.columns || [],
        rows: cell.rows || [],
        rows_affected: cell.rows_affected,
        duration: cell.duration,
      } as QueryOutput;
    }
    return $storeOutput;
  });

  const nextPage = () => {
    if (currentOutput && currentOutput.rows) {
      if (+limit + offset < currentOutput.rows.length) {
        offset = Math.min(+limit + offset, currentOutput.rows.length);
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
        exportToCSV(currentOutput, `output_${dateExt}.csv`);
        break;
      case "json":
        exportToJSON(outputToData(currentOutput), `output_${dateExt}.json`);
        break;
    }
  };

  const openInNewTab = () => {
    let htmlText = tableHTML(currentOutput);

    // Create a blob of the data
    const blob = new Blob([htmlText], { type: "text/html" });
    const url = URL.createObjectURL(blob);
    window.open(url, "_blank")?.focus();
  };

  const clearAll = () => {
    storeOutput.set(null);
    storeNotebookOutput.set(null);
    selectedCell = 0;
  };

  // Reset offset when output changes
  $effect(() => {
    if (currentOutput) {
      offset = 0;
    }
  });

  // Reset selectedCell when notebook output changes
  $effect(() => {
    if ($storeNotebookOutput) {
      selectedCell = 0;
    }
  });
</script>

<div class="grid">
  <div class="flex items-center bg-gray-300 justify-between">
    <div class="px-2 flex items-center">
      <span class="text-sm font-semibold">Query Output</span>
      {#if $storeNotebookOutput && $storeNotebookOutput.length > 0}
        <select
          class="ml-2 mr-1 border-none rounded-none bg-gray-300 hover:cursor-pointer hover:bg-gray-100 h-6 max-w-48 text-ellipsis text-xs"
          bind:value={selectedCell}
          title={$storeNotebookOutput[selectedCell]?.description || `Cell ${selectedCell + 1}`}
        >
          {#each $storeNotebookOutput as cell, i}
            <option value={i}>
              {i + 1}- {cell.description || `Cell ${i + 1}`} ({cell.status}{cell.error ? " - error" : ""})
            </option>
          {/each}
        </select>
      {/if}
      <select
        class="ml-2 mr-1 border-none rounded-none bg-gray-300 hover:cursor-pointer hover:bg-gray-100 w-28 h-6"
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
      {#if currentOutput?.rows_affected}
        <span class="text-xs text-gray-600 px-2">
          Rows Affected: {currentOutput.rows_affected}
        </span>
      {/if}
      {#if currentOutput?.duration}
        <span class="text-xs text-gray-600 px-2">
          Duration: {currentOutput.duration}
        </span>
      {/if}
      {#if currentOutput}
        <span class="text-xs text-gray-600 px-2">
          Offset: {offset}, Limit:
          <input
            type="text"
            size={`${limit}`.length || 1}
            bind:value={limit}
          />, Total: {currentOutput?.rows?.length ?? 0}
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
        onclick={clearAll}
      >
        <X />
      </button>
    </div>
  </div>
  {#if $storeNotebookOutput && $storeNotebookOutput[selectedCell]?.error}
    <div class="bg-red-100 text-red-800 px-3 py-1 text-sm">
      {$storeNotebookOutput[selectedCell].error}
    </div>
  {/if}
  <Rows output={currentOutput} {offset} {limit} />
</div>
