<script lang="ts">
  import Output from "@/components/Output.svelte";
  import Query from "@/components/Query.svelte";
  import {
    ArrowDown,
    GripVertical,
    Plus,
    ArrowRight,
    Save,
    Play,
    Trash,
  } from "@lucide/svelte";
  import type { cell, notebook } from "@/helper/model";
  import { storeInfo, storeNoteIds, storeOutput } from "@/store/store";
  import { ulid } from "ulid";
  import { reorder, useSortable } from "@/helper/sort.svelte";
  import {
    requestNote,
    requestRunNotebook,
    requestNoteSave,
    requestNoteDelete,
  } from "@/helper/call";
  import { addToast } from "@/store/toast";
  import { push } from "svelte-spa-router";

  let { params } = $props<{ params: { id: string } }>();
  let notebookID = $state<string>(ulid());
  let cells = $state<cell[]>([]);
  let name = $state("");
  let path = $state("");

  let sortable = $state<HTMLElement | null>(null);

  const addCell = () => {
    const newCell: cell = {
      id: ulid(),
      db_type: $storeInfo?.databases?.[0] || "",
      limit: 100,
      content: "",
      description: "",
      collapsed: false,
      enabled: true,
      result: false,
    };

    let cellsSnapshot = $state.snapshot(cells) || [];
    cellsSnapshot.push(newCell);

    cells = cellsSnapshot;
  };

  const removeCell = (id: string) => {
    cells = cells.filter((cell) => cell.id !== id);
  };

  const playNotebook = (path: string) => {
    addToast("Running notebook...", "info");
    storeOutput.set(null);
    requestRunNotebook(path)
      .then((response) => {
        storeOutput.set(null);
        addToast("Notebook run successfully", "info");
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

  const saveNotebook = () => {
    let savedCells: cell[] = [];
    cells.forEach((cell) => {
      savedCells.push({
        id: cell.id,
        db_type: cell.db_type,
        content: cell.content,
        limit: cell.limit,
        mode: cell.mode,
        description: cell.description,
        collapsed: cell.collapsed,
        enabled: cell.enabled,
        result: cell.result,
      });
    });
    const note: notebook = {
      id: notebookID,
      name: name,
      path: path,
      content: {
        cells: savedCells,
      },
    };
    console.log("Saving notebook:", note);

    requestNoteSave(note);
  };

  useSortable(() => sortable, {
    animation: 200,
    handle: ".sort-handle",
    onEnd(evt: any) {
      cells = reorder(cells, evt);
    },
  });

  const updateNotebook = async () => {
    // find the notebook by ID
    let notebook = await requestNote(params.id);
    if (notebook) {
      notebookID = notebook.id;
      name = notebook.name;
      path = notebook.path;
      cells = notebook.content.cells || [];
    } else {
      // if no notebook found, create a new one
      notebookID = params.id;
      name = "";
      path = "";
      cells = [];
    }
  };

  const deleteNotebook = (id: string) => {
    if (confirm("Are you sure you want to delete this notebook?")) {
      // Call the API to delete the notebook
      requestNoteDelete(id)
        .then(() => {
          // Remove the notebook from the store
          storeNoteIds.update((ids) => ids.filter((note) => note.id !== id));
          push(`/`);
          addToast("Notebook deleted successfully", "info");
        })
        .catch((error) => {
          addToast("Error deleting notebook: " + error.message, "alert");
        });
    }
  };

  $effect(() => {
    if (params.id) {
      // if exist than update the notebook
      if ($storeNoteIds.some((note) => note.id === params.id)) {
        updateNotebook();

        return;
      }

      notebookID = params.id;
      name = "";
      path = "";
      cells = [];
    }
  });
</script>

<div class="grid grid-rows-[1fr_auto] h-full w-full overflow-y-auto">
  <div class="flex flex-col h-full w-full min-h-0">
    <div class="border-b border-black flex justify-between">
      <input
        type="text"
        class="w-full px-2 py-1 hover:bg-white"
        bind:value={name}
        placeholder="Untitled Notebook"
      />
      <div class="flex gap-1 justify-between">
        <input
          type="text"
          class="px-2 py-1 hover:bg-white"
          bind:value={path}
          placeholder="path-to-notebook"
        />
        <button
          class="text-black px-2 py-1 hover:cursor-pointer hover:bg-red-500 hover:text-white flex gap-1"
          onclick={() => playNotebook(path)}
          title="Before to play need to save first"
        >
          <Play />
        </button>
        <button
          class="text-black px-2 py-1 hover:cursor-pointer hover:bg-blue-500 hover:text-white flex gap-1"
          onclick={saveNotebook}
        >
          <Save />
        </button>
        <button
          class="text-black px-2 py-1 hover:cursor-pointer hover:bg-red-500 hover:text-white flex gap-1"
          onclick={() => deleteNotebook(notebookID)}
        >
          <Trash />
        </button>
      </div>
    </div>

    <div class="relative flex flex-col h-full w-full overflow-y-auto min-h-0">
      <div bind:this={sortable}>
        {#each cells as cell, index (cell)}
          <div class="border-b border-gray-300 flex flex-row w-full">
            <div class="flex gap-1">
              <button
                class="sort-handle p-1 text-gray-500 hover:bg-yellow-200 hover:cursor-move"
              >
                <GripVertical />
              </button>
              <div class="flex items-center flex-col">
                <span>{index + 1}</span>
                <button
                  class="text-gray-500 hover:bg-gray-200 hover:cursor-pointer flex-1"
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
            </div>
            <Query
              bind:cell={cells[index]}
              deleteFunc={() => removeCell(cell.id)}
            />
          </div>
        {/each}
      </div>
      <button
        class="text-black px-2 py-1 mt-2 flex w-full border-t border-b border-black hover:bg-yellow-200 hover:cursor-pointer"
        onclick={addCell}
        title="Add a new cell"
      >
        <Plus class="text-blue-400" />
        <span class="ml-2">Add Cell</span>
      </button>
    </div>
  </div>

  <div class="border-t border-black">
    <Output />
  </div>
</div>
