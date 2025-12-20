<script lang="ts">
  import { storeInfo } from "@/store/store";
</script>

<div class="flex flex-col justify-between bg-gray-200">
  <div class="p-2">
    <span class="font-semibold text-black"> SAZ (Structured Access Zone) </span>
    <div class="divider my-1"></div>
    <div>
      <span>Connected Databases</span>
      <div class="mt-2">
        {#each ($storeInfo?.databases || []).sort( (a, b) => a.localeCompare(b), ) as db}
          <div
            class="px-1 bg-yellow-50 border border-black hover:shadow shadow-black"
          >
            {db}
          </div>
        {/each}
      </div>
    </div>
  </div>

  <div class="flex w-full justify-end border-t border-gray-300">
    <span>Version</span>
    <span class="px-1">
      {$storeInfo?.version}
    </span>
  </div>
</div>
