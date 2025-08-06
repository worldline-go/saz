<script lang="ts">
  import Name from "@/components/Name.svelte";
  import Output from "@/components/Output.svelte";
  import Query from "@/components/Query.svelte";
  import { Plus } from "@lucide/svelte";
  import type { Cell } from "@/helper/model";
  import { storeDatabases } from "@/store/store";
  import { ulid } from "ulid";

  let cells = $state<Cell[]>([]);

  const addCell = () => {
    const newCell: Cell = {
      id: ulid(),
      db_type: $storeDatabases[0] || "",
      content: "",
    };

    cells.push(newCell);
  };

  const removeCell = (id: string) => {
    cells = cells.filter((cell) => cell.id !== id);
  };
</script>

<div class="grid grid-cols-1 h-full w-full">
  <div class="relative flex flex-col h-full w-full">
    <div class="border-b border-gray-300 mb-1">
      <Name />
    </div>

    <div class="flex flex-col h-full w-full py-2">
      {#each cells as cell, index (cell.id)}
        <div class="border-b border-gray-300">
          <Query
            bind:query={cell.content}
            bind:db={cell.db_type}
            deleteFunc={() => removeCell(cell.id)}
          />
        </div>
      {/each}
      <button
        class="text-black px-2 py-1 mt-2 flex w-full border-t border-b border-gray-300 hover:bg-blue-50 hover:cursor-pointer"
        onclick={addCell}
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
