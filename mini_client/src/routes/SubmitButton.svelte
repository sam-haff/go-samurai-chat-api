<script lang="ts">
    import { fade } from 'svelte/transition';
    import { Button, Spinner } from 'flowbite-svelte'

    let { loading = $bindable(false), label='' } = $props();
    let uiLoading = $state(false);
</script>


<Button disabled={loading} type="submit" class="mt-4 mx-4 bg-primary-600 hover:bg-primary-700 active:bg-primary-750">
        {#if !loading}
            <div transition:fade={{duration: 200}} onoutroend={()=>{ uiLoading=true; }}>
                <span>{label}</span>
            </div>
        {:else if uiLoading}
            <div transition:fade={{duration: 200}}>
                <Spinner size='4' />
            </div>
        {/if}
</Button>