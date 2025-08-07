<script lang="ts">
  import Output from "@/components/Output.svelte";
  import Query from "@/components/Query.svelte";
  import {
    ArrowDown,
    GripVertical,
    Plus,
    ArrowRight,
    Save,
  } from "@lucide/svelte";
  import type { cell, notebook } from "@/helper/model";
  import { storeInfo } from "@/store/store";
  import { ulid } from "ulid";
  import { reorder, useSortable } from "@/helper/sort.svelte";

  let cells = $state<cell[]>([]);
  let name = $state("");

  let sortable = $state<HTMLElement | null>(null);

  const addCell = () => {
    const newCell: cell = {
      id: ulid(),
      db_type: $storeInfo?.databases?.[0] || "",
      content: "",
      description: "",
      collapsed: false,
    };

    cells.push(newCell);
  };

  const removeCell = (id: string) => {
    cells = cells.filter((cell) => cell.id !== id);
  };

  const saveNotebook = () => {
    let savedCells: cell[] = [];
    cells.forEach((cell) => {
      savedCells.push({
        id: cell.id,
        db_type: cell.db_type,
        content: cell.content,
        description: cell.description,
        collapsed: cell.collapsed,
      });
    });
    const note: notebook = {
      name: name,
      cells: savedCells,
    };
    console.log("Saving notebook:", note);
  };

  useSortable(() => sortable, {
    animation: 200,
    onEnd(evt: any) {
      cells = reorder(cells, evt);
    },
  });
</script>

<div class="grid grid-rows-[1fr_auto] h-full w-full overflow-y-auto">
  <div class="flex flex-col h-full w-full min-h-0">
    <div class="border-b border-gray-300 mb-1 flex justify-between">
      <input
        type="text"
        class="w-full px-2 py-1"
        bind:value={name}
        placeholder="Untitled Notebook"
      />

      <button
        class="text-black px-2 py-1 hover:cursor-pointer hover:bg-blue-500 hover:text-white flex gap-1"
        onclick={saveNotebook}
      >
        <Save />
        <span>Save</span>
      </button>
    </div>

    <div
      class="relative flex flex-col h-full w-full py-2 overflow-y-auto min-h-0"
    >
      <div bind:this={sortable}>
        {#each cells as cell (cell)}
          <div class="border-b border-gray-300 flex flex-row w-full">
            <div class="flex gap-1">
              <button
                class="p-1 text-gray-500 hover:bg-gray-200 hover:cursor-move"
              >
                <GripVertical />
              </button>
              <button
                class="text-gray-500 hover:bg-gray-200 hover:cursor-pointer"
                onclick={() => {
                  cell.collapsed = !cell.collapsed;
                }}
              >
                {#if cell.collapsed}
                  <ArrowRight />
                {:else}
                  <ArrowDown />
                {/if}
              </button>
            </div>
            <Query
              bind:query={cell.content}
              bind:db={cell.db_type}
              deleteFunc={() => removeCell(cell.id)}
              bind:collapsed={cell.collapsed}
              bind:description={cell.description}
            />
          </div>
        {/each}
      </div>
      <button
        class="text-black px-2 py-1 mt-2 flex w-full border-t border-b border-gray-300 hover:bg-blue-50 hover:cursor-pointer"
        onclick={addCell}
        title="Add a new cell"
      >
        <Plus class="text-blue-400" />
        <span class="ml-2">Add Cell</span>
      </button>
    </div>
  </div>

  <div class="border-t border-gray-300 px-1">
    <Output />
  </div>
</div>
