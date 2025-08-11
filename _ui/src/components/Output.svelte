<script lang="ts">
  import { ArrowLeft, ArrowRight, X } from "@lucide/svelte";
  import Rows from "./Rows.svelte";
  import { storeOutput } from "@/store/store";

  let offset = $state(0);
  let limit = $state(10);

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

  $effect(() => {
    if ($storeOutput) {
      offset = 0;
    }
  });
</script>

<div>
  <div class="flex items-center bg-gray-300 justify-between">
    <span class="px-2 text-sm font-semibold">Query Output</span>
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
