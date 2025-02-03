<script lang='ts'>
    import { cubicIn } from "svelte/easing";
    import { fade, scale } from "svelte/transition";

	let { showModal = $bindable(), header, children, preventModalClose } = $props();

    let dialog = $state();

</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
{#if showModal}
 <div 
    in:fade={{duration: 100}}
    aria-roledescription="background"
    onclick={(event: any) => {
        console.log(event);
        if (event.pointerType === "") { return;}
        console.log("Modal click!");
        if (preventModalClose) { return; } 
        let dialogEl = dialog as HTMLElement;
        let rect = dialogEl.getBoundingClientRect();
        if(event.clientY < rect.top || event.clientY > rect.bottom) { showModal = false; }
        if(event.clientX < rect.left || event.clientX > rect.right) { showModal = false; }
    }}
    class="h-screen w-screen absolute z-40 left-0 top-0 bg-black/30"
    >
    <div in:scale={{easing:cubicIn, duration:250, start:0.95}} bind:this={dialog}  class="dialog-div dialog-open relative rounded shadow-sm mx-auto !z-40 bg-gray-100" style="">
        {@render children()}
    </div>
</div>
{/if}

<style>
	.dialog-div {
        max-width: 32em;
        width: 60vw;
        max-height: 90vh;
        min-height: 30vh;
        height: 60vh;
		border-radius: 1.2em;
		border: none;
		padding: 0;
        margin-top: 10vh;
	}
@media (max-width: 600px) {
    .dialog-div {
        width: 80vw;
    }
}
	.dialog-div > div {
		padding: 1em;
	}
	.dialog-div .dialog-open {
		animation: zoom 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
	}
	@keyframes zoom {
		from {
			transform: scale(0.95);
		}
		to {
			transform: scale(1);
		}
	}
	button {
		display: block;
	}
</style>