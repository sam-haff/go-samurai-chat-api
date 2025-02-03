<script>
	let { showModal = $bindable(), header, children } = $props();

	let dialog = $state(); // HTMLDialogElement

	$effect(() => {
		if (showModal) {
            dialog.showModal()
        } else {
            dialog.close();
        }
	});
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
<dialog
	bind:this={dialog}
	onclose={() => (showModal = false)}
    onclick={(event) => { 
        let rect = dialog.getBoundingClientRect();
        if(event.clientY < rect.top || event.clientY > rect.bottom) { showModal = false; return dialog.close(); }
        if(event.clientX < rect.left || event.clientX > rect.right) { showModal = false; return dialog.close();}
    }}
>
		{@render children?.()}
</dialog>

<style>
	dialog {
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
    dialog {
        width: 80vw;
    }
}
	dialog::backdrop {
		background: rgba(0, 0, 0, 0.3);
	}
	dialog > div {
		padding: 1em;
	}
	dialog[open] {
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
	dialog[open]::backdrop {
		animation: fade 0.2s ease-out;
	}
	@keyframes fade {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
	button {
		display: block;
	}
</style>