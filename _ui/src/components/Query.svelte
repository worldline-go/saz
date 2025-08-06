<script lang="ts">
  import Editor from "@/components/Editor.svelte";
  import { requestRun } from "@/helper/call";
  import { storeDatabases, storeOutput } from "@/store/store";
  import { addToast } from "@/store/toast";
  import { Scan, Ellipsis, Trash } from "@lucide/svelte";

  let {
    query = $bindable(),
    db = $bindable(),
    deleteFunc = $bindable(),
  } = $props();

  const runQuery = () => {
    requestRun({ name: db, query: query })
      .then((response) => {
        storeOutput.set(response.data?.data || []);
      })
      .catch((error) => {
        storeOutput.set([]);
        addToast("Error running query: " + error.message, "alert");
      });
  };

  let fullScreen = $state(false);
</script>

<div
  class={[
    "grid grid-flow-col grid-cols-1 bg-gray-100 w-full",
    fullScreen
      ? "absolute top-0 left-0 h-auto w-full bg-white z-10"
      : "hover:bg-white focus:bg-white",
  ]}
>
  <div>
    <div class="flex justify-between border-b border-gray-300 pb-1">
      <select
        class="select border-none rounded-none hover:cursor-pointer hover:bg-white px-2 py-1 w-28"
        bind:value={db}
      >
        {#each $storeDatabases as database}
          <option value={database}>{database}</option>
        {/each}
      </select>
      <button
        class="text-black py-1 px-2 hover:cursor-pointer hover:bg-red-500 hover:text-white"
        onclick={runQuery}
      >
        Run Query
      </button>
    </div>
    <div class="overflow-y-auto ov">
      <Editor bind:value={query} />
    </div>
  </div>

  <div class="flex flex-col items-center pl-1 gap-1">
    <button
      class="p-1 text-white hover:bg-gray-300 hover:cursor-pointer"
      onclick={() => (fullScreen = !fullScreen)}
    >
      <Scan class="text-gray-600" />
    </button>

    <details class="dropdown dropdown-bottom dropdown-end marker:content-['']">
      <summary class="p-1 text-white hover:bg-gray-300 hover:cursor-pointer">
        <Ellipsis class="text-gray-600" />
      </summary>
      <ul class="dropdown-content bg-base-100 z-1 shadow-sm">
        <li>
          <button
            class="p-1 text-black hover:bg-gray-300 hover:cursor-pointer flex text-sm items-center"
            onclick={deleteFunc}
          >
            <Trash class="p-1" />
            <span class="ml-1">Delete</span>
          </button>
        </li>
      </ul>
    </details>
  </div>
</div>
