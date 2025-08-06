<script lang="ts">
  import Rows from "./Rows.svelte";
  import { storeOutput } from "@/store/store";
</script>

<div>
  <Rows rows={$storeOutput} />
</div>
