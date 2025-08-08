<script lang="ts">
  import Editor from "@/components/Editor.svelte";
  import { requestRun } from "@/helper/call";
  import { storeInfo, storeOutput } from "@/store/store";
  import { addToast } from "@/store/toast";
  import {
    Scan,
    Ellipsis,
    Trash,
    TrainTrack,
    BotOff,
    Play,
    Wifi,
    WifiOff,
    Captions,
    CaptionsOff,
    BrushCleaning,
    Route,
    RouteOff,
    ShieldPlus,
  } from "@lucide/svelte";
  import type { cell as cellType } from "@/helper/model";

  let {
    deleteFunc = $bindable(),
    cell = $bindable<cellType>(),
  }: { deleteFunc: () => void; cell: cellType } = $props();

  const runQuery = () => {
    requestRun(cell)
      .then((response) => {
        storeOutput.set(response.data);
      })
      .catch((error) => {
        storeOutput.set(null);
        addToast("Error running query: " + error.message, "alert");
      });
  };

  let fullScreen = $state(false);
  let dropdownRef = $state<HTMLDetailsElement | null>(null);

  const setMode = (m: "transfer" | null) => {
    switch (m) {
      case "transfer":
        cell.mode = {
          enabled: true,
          name: "transfer",
          db_type: cell.db_type,
          table: "",
          wipe: false,
        };
        break;
      default:
        cell.mode = undefined;
    }
  };
</script>

<div
  class={[
    "grid grid-flow-col grid-cols-1 bg-gray-100 w-full text-sm",
    fullScreen
      ? "absolute top-0 left-0 h-full w-full bg-gray-50 z-10"
      : "hover:bg-gray-50",
  ]}
>
  <div class="border-l border-r border-gray-300">
    <div class="flex justify-between border-b border-gray-300">
      <div class="flex items-center w-full">
        <select
          class="select border-none rounded-none bg-gray-100 hover:cursor-pointer hover:bg-white px-2 py-1 w-28 h-7"
          bind:value={cell.db_type}
        >
          {#each $storeInfo?.databases ?? [] as database}
            <option value={database}>{database}</option>
          {/each}
        </select>
        <span class="divider divider-horizontal mx-0 !w-[1px]"></span>
        <input
          class="input h-full border-none rounded-none bg-gray-100 hover:cursor-text hover:bg-white focus:bg-white px-2 w-full"
          type="text"
          placeholder="Describe your query"
          bind:value={cell.description}
        />
      </div>
      <div class="flex items-center gap-1">
        <label
          class="swap hover:bg-yellow-300 hover:cursor-pointer px-2 h-full"
          title="Show Output Result"
        >
          <input type="checkbox" bind:checked={cell.result} />
          <div class="swap-on"><Captions /></div>
          <div class="swap-off"><CaptionsOff /></div>
        </label>
        <label
          class="swap hover:bg-yellow-300 hover:cursor-pointer px-2 h-full"
          title="Enable/Disable Query on the notebook execution"
        >
          <input type="checkbox" bind:checked={cell.enabled} />
          <div class="swap-on"><Wifi /></div>
          <div class="swap-off"><WifiOff /></div>
        </label>
        <button
          class="text-gray-600 px-2 hover:cursor-pointer hover:bg-red-500 hover:text-white h-full"
          onclick={runQuery}
          title="Run Query"
        >
          <Play />
        </button>
      </div>
    </div>
    <div class="flex justify-between">
      <div class="overflow-y-auto w-full">
        <Editor bind:value={cell.content} collapse={cell.collapsed} />
      </div>
      {#if cell.mode?.name === "transfer"}
        <div class="flex flex-col items-start border-l border-gray-300">
          <span
            class={[
              "w-48 border-b border-gray-300 flex items-center justify-between",
              cell.mode.enabled ? "bg-green-100" : "bg-red-100",
            ]}
          >
            <span class="px-2">Transfer Mode</span>
            <div>
              <label class="swap hover:bg-yellow-300">
                <input type="checkbox" bind:checked={cell.mode.wipe} />
                <div class="swap-on" title="Append to existing data">
                  <BrushCleaning class="px-1" />
                </div>
                <div class="swap-off" title="Wipe before transfer">
                  <ShieldPlus class="px-1" />
                </div>
              </label>
              <label class="swap hover:bg-yellow-300">
                <input type="checkbox" bind:checked={cell.mode.enabled} />
                <div class="swap-on" title="Disable Mode">
                  <Route class="px-1" />
                </div>
                <div class="swap-off" title="Enable Mode">
                  <RouteOff class="px-1" />
                </div>
              </label>
            </div>
          </span>
          <select
            class="select border-none rounded-none hover:cursor-pointer hover:bg-white px-2 h-7"
            bind:value={cell.mode.db_type}
          >
            {#each $storeInfo?.databases ?? [] as database}
              <option value={database}>{database}</option>
            {/each}
          </select>
          <input
            class="bg-white border-b border-t border-gray-300 hover:cursor-text hover:bg-white px-2 py-1 w-full"
            type="text"
            placeholder="Table Name"
            bind:value={cell.mode.table}
          />
        </div>
      {/if}
    </div>
  </div>

  <div class="flex flex-col items-center gap-1">
    <button
      class="p-1 text-white hover:bg-gray-300 hover:cursor-pointer"
      onclick={() => (fullScreen = !fullScreen)}
      title="Toggle Fullscreen"
    >
      <Scan class="text-gray-600" />
    </button>

    <details
      bind:this={dropdownRef}
      class="dropdown dropdown-bottom dropdown-end marker:content-['']"
    >
      <summary
        class="p-1 text-white hover:bg-gray-300 hover:cursor-pointer"
        title="More options"
      >
        <Ellipsis class="text-gray-600" />
      </summary>
      <ul class="dropdown-content bg-base-100 z-1 shadow-sm w-32">
        <li>
          <button
            class="w-full p-1 text-black hover:bg-blue-500 hover:text-white hover:cursor-pointer flex text-sm items-center"
            onclick={() => {
              setMode(null);
              dropdownRef?.removeAttribute("open");
            }}
          >
            <BotOff class="p-1" />
            <span class="ml-1">Disable Mode</span>
          </button>
        </li>
        <li>
          <button
            class="w-full p-1 text-black hover:bg-blue-500 hover:text-white hover:cursor-pointer flex text-sm items-center"
            onclick={() => {
              setMode("transfer");
              dropdownRef?.removeAttribute("open");
            }}
          >
            <TrainTrack class="p-1" />
            <span class="ml-1">Transfer Mode</span>
          </button>
        </li>
        <li>
          <span class="divider my-0"></span>
          <button
            class="w-full p-1 text-black hover:bg-red-500 hover:text-white hover:cursor-pointer flex text-sm items-center"
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
