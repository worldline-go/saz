<script lang="ts">
  import { onMount } from "svelte";
  // import type { ComponentType } from "svelte";
  type RouteComponent = typeof MainPage;
  import Router from "svelte-spa-router";

  import { storeNavbar } from "@/store/store";
  import { addToast } from "@/store/toast";

  import Sidebar from "@/components/Sidebar.svelte";
  import Toast from "./components/Toast.svelte";

  import MainPage from "@/pages/Main.svelte";

  import { requestInfo } from "@/helper/call";

  let layout: HTMLElement;
  let mounted = $state(false);

  const routes = new Map<string | RegExp, RouteComponent>();
  routes.set("/*", MainPage);

  onMount(async () => {
    await requestInfo();
    mounted = true;
  });
</script>

<Toast />

<div
  bind:this={layout}
  class={[
    "grid grid-flow-col h-full w-full relative overflow-y-auto bg-slate-100",
    $storeNavbar.sideBarOpen ? "grid-cols-[8rem]" : "grid-cols-[0]",
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
