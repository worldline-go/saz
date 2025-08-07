<script lang="ts">
  import Editor from "@/components/Editor.svelte";
  import { requestRun } from "@/helper/call";
  import { storeInfo, storeOutput } from "@/store/store";
  import { addToast } from "@/store/toast";
  import { Scan, Ellipsis, Trash } from "@lucide/svelte";

  let {
    query = $bindable(),
    description = $bindable(),
    db = $bindable(),
    deleteFunc = $bindable(),
    collapsed = $bindable(false),
  } = $props();

  const runQuery = () => {
    requestRun({ name: db, query: query })
      .then((response) => {
        storeOutput.set(response.data);
      })
      .catch((error) => {
        storeOutput.set(null);
        addToast("Error running query: " + error.message, "alert");
      });
  };

  let fullScreen = $state(false);
</script>

<div
  class={[
    "grid grid-flow-col grid-cols-1 bg-gray-100 w-full",
    fullScreen
      ? "absolute top-0 left-0 h-full w-full bg-gray-50 z-10"
      : "hover:bg-gray-50",
  ]}
>
  <div>
    <div class="flex justify-between border-b border-gray-300 pb-1">
      <div>
        <select
          class="select border-none rounded-none bg-gray-100 hover:cursor-pointer hover:bg-white px-2 py-1 w-28"
          bind:value={db}
        >
          {#each $storeInfo?.databases ?? [] as database}
            <option value={database}>{database}</option>
          {/each}
        </select>
        <input
          class="input border-none rounded-none bg-gray-100 hover:cursor-pointer hover:bg-white focus:bg-white px-2 py-1 w-64"
          type="text"
          placeholder="Describe your query"
          bind:value={description}
        />
      </div>
      <button
        class="text-black py-1 px-2 hover:cursor-pointer hover:bg-red-500 hover:text-white"
        onclick={runQuery}
      >
        Run Query
      </button>
    </div>
    <div class="overflow-y-auto">
      <Editor bind:value={query} collapse={collapsed} />
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
