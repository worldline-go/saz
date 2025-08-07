<script lang="ts">
  import { ArrowLeft } from "@lucide/svelte";
  import Rows from "./Rows.svelte";
  import { storeOutput } from "@/store/store";
</script>

<div>
  <div class="flex items-center bg-green-100">
    <span class="px-2 text-sm font-semibold">Query Output</span>
    <div>
      <button class=" text-gray-500 hover:bg-gray-200 hover:cursor-pointer">
        <ArrowLeft />
      </button>
    </div>
  </div>
  <Rows output={$storeOutput} />
</div>
