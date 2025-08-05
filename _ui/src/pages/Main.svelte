<script lang="ts">
  import Editor from "@/components/Editor.svelte";
  import Rows from "@/components/Rows.svelte";
  import { requestRun } from "@/helper/call";
  import { storeNavbar } from "@/store/store";
  import update from "immutability-helper";

  storeNavbar.update((v) => update(v, { title: { $set: "Main" } }));

  let queryValue = $state("");
  let rows: Record<string, any>[] = $state([]);

  const runQuery = () => {
    requestRun({ name: "postgres", query: queryValue })
      .then((response) => {
        rows = response.data?.data || [];
        console.log("Query response:", response.data);
      })
      .catch((error) => {
        console.error("Error running query:", error);
      });
  };
</script>

<div>
  <div>
    <select class="m-2 p-2">
      <option value="postgres">Postgres</option>
      <option value="mysql">MySQL</option>
      <option value="sqlite">SQLite</option>
    </select>
    <button
      class="bg-yellow-400 text-black p-2 m-2 hover:cursor-pointer hover:bg-red-500 hover:text-white"
      onclick={runQuery}
    >
      Run Query
    </button>
  </div>
  <Editor bind:value={queryValue} />
  <Rows bind:rows />
</div>
