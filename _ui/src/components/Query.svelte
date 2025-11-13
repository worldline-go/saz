<script lang="ts">
  import Editor from "@/components/Editor.svelte";
  import { requestRun, requestRunTemplate } from "@/helper/call";
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
    MapPlus,
    SquareDashed,
    Plus,
    Vote,
    RefreshCw,
    RefreshCwOff,
    Dna,
    DnaOff,
    BookMarked,
    BookDashed,
    Bot,
    ZapOff,
    Zap,
    ListVideo,
  } from "@lucide/svelte";
  import { encodingTypes, type cell as cellType } from "@/helper/model";

  let {
    deleteFunc = $bindable(),
    cell = $bindable<cellType>(),
  }: { deleteFunc: () => void; cell: cellType } = $props();

  let preview = $state(false);
  let previewResult = $state("");

  let clearPreview = () => {
    previewResult = "";
  };

  $effect(() => {
    if (!preview) {
      clearPreview();
    } else {
      runTemplate();
    }
  });

  const runQuery = () => {
    addToast("Running cell...", "info");
    storeOutput.set(null);
    requestRun(cell)
      .then((response) => {
        storeOutput.set(response.data);
        addToast("Cell run successfully", "info");
      })
      .catch((error) => {
        if (error.response) {
          storeOutput.set({
            columns: ["message", "error"],
            rows: [
              [error.response?.data?.message, error.response?.data?.error],
            ],
          });
        } else {
          storeOutput.set(null);
        }
        addToast("Error running query: " + error.message, "alert");
      });
  };

  const runTemplate = () => {
    requestRunTemplate(cell.content, {})
      .then((response) => {
        previewResult = response.data?.data || "";
      })
      .catch((error) => {
        previewResult = error?.response?.data?.error;
        addToast(error?.response?.data?.error, "alert");
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
          skip_error: {
            enabled: false,
            message: "",
          },
          map_type: {
            enabled: false,
            column: {},
            destination: {},
          },
          batch: 1,
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
          class="select border-none rounded-none bg-gray-100 hover:cursor-pointer hover:bg-white pl-2 pr-0 w-28 h-7"
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
        {#if cell.template?.enabled}
          <label
            class="swap hover:bg-yellow-200 hover:cursor-pointer px-2 h-full"
            title="Enable/Disable Preview Mode"
          >
            <input type="checkbox" bind:checked={preview} />
            <div class="swap-on"><Zap /></div>
            <div class="swap-off"><ZapOff /></div>
          </label>
        {/if}
        {#if preview}
          <button
            class="px-2 h-full text-black hover:bg-green-400 hover:cursor-pointer"
            onclick={runTemplate}
            title="Render Template"
          >
            <ListVideo />
          </button>
        {/if}
        <input
          class={[
            "h-full border-none rounded-none bg-gray-100 hover:cursor-text px-2 w-20",
            cell.result ? "hover:bg-white focus:bg-white" : "line-through",
          ]}
          disabled={!cell.result}
          type="number"
          placeholder="Limit"
          title="Total Limit"
          bind:value={cell.limit}
        />
        <label
          class="swap hover:bg-yellow-200 hover:cursor-pointer px-2 h-full"
          title="Show Output Result"
        >
          <input type="checkbox" bind:checked={cell.result} />
          <div class="swap-on"><Captions /></div>
          <div class="swap-off"><CaptionsOff /></div>
        </label>
        <label
          class="swap hover:bg-yellow-200 hover:cursor-pointer px-2 h-full"
          title="Enable/Disable template in query"
        >
          <input type="checkbox" bind:checked={cell.template.enabled} />
          <div class="swap-on"><Bot /></div>
          <div class="swap-off"><BotOff /></div>
        </label>
        <label
          class="swap hover:bg-yellow-200 hover:cursor-pointer px-2 h-full"
          title="Enable/Disable Query on the notebook execution"
        >
          <input type="checkbox" bind:checked={cell.enabled} />
          <div class="swap-on"><Wifi /></div>
          <div class="swap-off"><WifiOff /></div>
        </label>
        <button
          class="text-black px-2 hover:cursor-pointer hover:bg-red-500 hover:text-white h-full"
          onclick={runQuery}
          title="Run Query"
        >
          <Play />
        </button>
      </div>
    </div>
    <div class="flex justify-between">
      <div class="overflow-y-auto w-full">
        <Editor
          bind:value={cell.content}
          collapse={cell.collapsed}
          class="bg-white"
        />
      </div>
      {#if preview}
        <div class="min-w-56 w-full">
          <Editor
            bind:value={previewResult}
            readonly={true}
            class="bg-yellow-50"
          />
        </div>
      {/if}
      {#if cell.mode?.name === "transfer"}
        {#if cell.mode?.map_type.enabled}
          <div class="flex flex-col items-start border-l border-gray-300">
            <span
              class={"w-full min-w-56 border-b border-gray-300 flex items-center justify-between leading-[normal]"}
            >
              <span class="px-2">Source</span>
              <button
                class="hover:bg-yellow-200 hover:cursor-pointer"
                title="Add new column"
                onclick={() => {
                  if (cell?.mode) {
                    if (!cell.mode.map_type.column) {
                      cell.mode.map_type.column = {};
                    }

                    cell.mode.map_type.column[""] = {
                      type: "string",
                      nullable: false,
                    };
                  }
                }}
              >
                <Plus class="px-1" />
              </button>
            </span>
            {#if cell.mode?.map_type.column}
              <div class="w-full border-l-4 border-teal-300">
                {#each Object.entries(cell.mode.map_type.column) as [columnName, column]}
                  <div
                    class="flex items-center justify-between border-b border-gray-300 w-full"
                  >
                    <input
                      class="px-2 hover:bg-white h-full"
                      value={columnName}
                      placeholder="Untitled"
                      onchange={(e) => {
                        const newName = e.currentTarget.value.trim();
                        if (cell.mode?.map_type.column && newName) {
                          cell.mode.map_type.column[newName] =
                            cell.mode.map_type.column[columnName];
                          delete cell.mode.map_type.column[columnName];
                        }
                      }}
                    />
                    <div class="flex items-center">
                      <select
                        class="appearance-none bg-none border-none rounded-none hover:cursor-pointer hover:bg-white pl-1 pr-0 h-7"
                        bind:value={column.type}
                      >
                        <option value="string">String</option>
                        <option value="number">Number</option>
                        <option value="date">Date</option>
                      </select>
                      <label class="swap hover:bg-yellow-200 h-full">
                        <input type="checkbox" bind:checked={column.nullable} />
                        <div class="swap-on" title="Not Null">
                          <SquareDashed class="px-1" />
                        </div>
                        <div class="swap-off" title="Null">
                          <Vote class="px-1" />
                        </div>
                      </label>
                      <button
                        class="h-full hover:cursor-pointer hover:bg-red-500 hover:text-white"
                        onclick={() => {
                          delete cell.mode?.map_type.column![columnName];
                        }}
                      >
                        <Trash class="px-1" />
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
            <span
              class={"w-full min-w-56 border-b border-gray-300 flex items-center justify-between leading-[normal]"}
            >
              <span class="px-2">Destination</span>
              <button
                class="hover:bg-yellow-200 hover:cursor-pointer"
                title="Add new column"
                onclick={() => {
                  if (cell?.mode) {
                    if (!cell.mode.map_type.destination) {
                      cell.mode.map_type.destination = {};
                    }

                    cell.mode.map_type.destination[""] = {
                      type: "string",
                      nullable: false,
                      template: {
                        enabled: false,
                        value: "",
                      },
                      encoding: {
                        enabled: false,
                        coding: "ISO 8859-1",
                      },
                    };
                  }
                }}
              >
                <Plus class="px-1" />
              </button>
            </span>
            {#if cell.mode?.map_type.destination}
              <div class="w-full border-l-4 border-blue-300">
                {#each Object.entries(cell.mode.map_type.destination) as [columnName, column]}
                  <div
                    class="flex items-center justify-between border-b border-gray-300 w-full"
                  >
                    <input
                      class="px-2 hover:bg-white h-full"
                      value={columnName}
                      placeholder="Untitled"
                      onchange={(e) => {
                        const newName = e.currentTarget.value.trim();
                        if (cell.mode?.map_type.destination && newName) {
                          cell.mode.map_type.destination[newName] =
                            cell.mode.map_type.destination[columnName];
                          delete cell.mode.map_type.destination[columnName];
                        }
                      }}
                    />
                    <div class="flex items-center">
                      <select
                        class="appearance-none bg-none border-none rounded-none hover:cursor-pointer hover:bg-white pl-1 pr-0 h-7"
                        bind:value={column.type}
                      >
                        <option value="string">String</option>
                        <option value="number">Number</option>
                        <option value="date">Date</option>
                      </select>
                      <label class="swap hover:bg-yellow-200 h-full">
                        <input
                          type="checkbox"
                          bind:checked={column.template.enabled}
                        />
                        <div class="swap-on" title="Template Disabled">
                          <RefreshCw class="px-1" />
                        </div>
                        <div class="swap-off" title="Template Enabled">
                          <RefreshCwOff class="px-1" />
                        </div>
                      </label>
                      <label class="swap hover:bg-yellow-200 h-full">
                        <input
                          type="checkbox"
                          bind:checked={column.encoding.enabled}
                        />
                        <div class="swap-on" title="Encoding Disabled">
                          <BookMarked class="px-1" />
                        </div>
                        <div class="swap-off" title="Encoding Enabled">
                          <BookDashed class="px-1" />
                        </div>
                      </label>
                      <label class="swap hover:bg-yellow-200 h-full">
                        <input type="checkbox" bind:checked={column.nullable} />
                        <div class="swap-on" title="Not Null">
                          <SquareDashed class="px-1" />
                        </div>
                        <div class="swap-off" title="Null">
                          <Vote class="px-1" />
                        </div>
                      </label>
                      <button
                        class="h-full hover:cursor-pointer hover:bg-red-500 hover:text-white"
                        onclick={() => {
                          delete cell.mode?.map_type.destination![columnName];
                        }}
                      >
                        <Trash class="px-1" />
                      </button>
                    </div>
                  </div>
                  {#if column.template.enabled}
                    <div class=" w-full border-b border-gray-300">
                      <input
                        class="px-2 py-1 hover:bg-white h-full w-full border-l-4 border-yellow-200"
                        type="text"
                        placeholder="Template Value"
                        bind:value={column.template.value}
                        disabled={!column.template.enabled}
                      />
                    </div>
                  {/if}
                  {#if column.encoding.enabled}
                    <div
                      class=" w-full border-b border-b-gray-300 border-l-4 border-l-pink-400 flex justify-between items-center"
                    >
                      <span class="px-2 py-1">Encoding:</span>
                      <select
                        class="select border-none rounded-none bg-gray-100 hover:cursor-pointer hover:bg-white pl-2 pr-0 w-28 h-7"
                        bind:value={column.encoding.coding}
                      >
                        {#each encodingTypes as coding}
                          <option value={coding}>{coding}</option>
                        {/each}
                      </select>
                    </div>
                  {/if}
                {/each}
              </div>
            {/if}
          </div>
        {/if}
        <div class="flex flex-col items-center border-l border-gray-300">
          <span
            class={[
              "w-48 border-b border-gray-300 flex items-center justify-between leading-[normal]",
              cell.mode.enabled ? "bg-green-100" : "bg-red-100",
            ]}
          >
            <span class="px-2">Transfer</span>
            <div>
              <label class="swap hover:bg-yellow-200">
                <input
                  type="checkbox"
                  bind:checked={cell.mode.map_type.enabled}
                />
                <div class="swap-on" title="Disable column mapping">
                  <MapPlus class="px-1" />
                </div>
                <div class="swap-off" title="Column Mapping">
                  <SquareDashed class="px-1" />
                </div>
              </label>
              <label class="swap hover:bg-yellow-200">
                <input type="checkbox" bind:checked={cell.mode.wipe} />
                <div class="swap-on" title="Append to existing data">
                  <BrushCleaning class="px-1" />
                </div>
                <div class="swap-off" title="Wipe before transfer">
                  <ShieldPlus class="px-1" />
                </div>
              </label>
              <label class="swap hover:bg-yellow-200">
                <input
                  type="checkbox"
                  bind:checked={cell.mode.skip_error.enabled}
                />
                <div class="swap-on" title="Disabled Skip Error">
                  <Dna class="px-1" />
                </div>
                <div class="swap-off" title="Skip Error">
                  <DnaOff class="px-1" />
                </div>
              </label>
              <label class="swap hover:bg-yellow-200">
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
            class="select border-none rounded-none hover:cursor-pointer hover:bg-white pl-2 pr-0 h-7"
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
          <input
            class="bg-white border-b border-t border-gray-300 hover:cursor-text hover:bg-white px-2 py-1 w-full"
            type="number"
            placeholder="1"
            min="1"
            title="Batch Size"
            bind:value={cell.mode.batch}
          />
          {#if cell.mode.skip_error.enabled}
            <input
              class="bg-yellow-100 border-b border-t border-gray-300 hover:cursor-text hover:bg-white px-2 py-1 w-full"
              type="text"
              placeholder="Error Message"
              bind:value={cell.mode.skip_error.message}
              disabled={!cell.mode.skip_error.enabled}
            />
          {/if}
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
