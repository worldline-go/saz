<script lang="ts">
  import { onMount } from "svelte";
  import Router from "svelte-spa-router";

  import { storeNavbar } from "@/store/store";
  import { addToast } from "@/store/toast";

  import Sidebar from "@/components/Sidebar.svelte";
  import Toast from "./components/Toast.svelte";

  import NotePage from "@/pages/Note.svelte";
  import MainPage from "@/pages/Main.svelte";

  import { requestInfo, requestNotes } from "@/helper/call";

  let layout: HTMLElement;
  let mounted = $state(false);

  const routes = {
    "/note/:id": NotePage,
    "/": MainPage,
  };

  onMount(async () => {
    await requestInfo();
    await requestNotes();
    mounted = true;
  });
</script>

<Toast />

<div
  bind:this={layout}
  class={[
    "grid grid-flow-col h-full w-full relative overflow-y-auto bg-slate-100",
    $storeNavbar.sideBarOpen ? "grid-cols-[12rem]" : "grid-cols-[0]",
  ]}
>
  {#if !mounted}
    <div class="absolute inset-0 flex items-center justify-center">
      <div
        class="animate-spin rounded-full h-32 w-32 border-t-2 border-b-2 border-gray-900"
      ></div>
    </div>
  {:else}
    <Sidebar />
    <Router {routes} />
  {/if}
</div>
