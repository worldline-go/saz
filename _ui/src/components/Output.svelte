<script lang="ts">
  import { ArrowLeft, ArrowRight, X } from "@lucide/svelte";
  import Rows from "./Rows.svelte";
  import { storeOutput } from "@/store/store";
</script>

<div>
  <div class="flex items-center bg-gray-300 justify-between">
    <span class="px-2 text-sm font-semibold">Query Output</span>
    <div class="flex items-center">
      <button class=" text-gray-500 hover:bg-gray-200 hover:cursor-pointer">
        <ArrowLeft />
      </button>
      <button class=" text-gray-500 hover:bg-gray-200 hover:cursor-pointer">
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
  <Rows output={$storeOutput} />
</div>
