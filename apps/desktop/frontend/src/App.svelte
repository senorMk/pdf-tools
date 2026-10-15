<script lang="ts">
  import { onMount } from 'svelte';
  import Icon from './Icon.svelte';
  import About from './About.svelte';
  import { appInfo } from './app-info';
  import { desktop, isDesktop, onFileDrop, type Document } from './bridge';

  let file: Document | null = null;
  let password = '';
  let showPassword = false;
  let busy = false;
  let error = '';
  let notice = '';
  let savedPath = '';
  let native = false;
  let passwordInput: HTMLInputElement;
  let page: 'password' | 'about' = 'password';

  function navigate(nextPage: typeof page) {
    page = nextPage;
    password = '';
    showPassword = false;
  }

  const message = (value: unknown) => value instanceof Error ? value.message : String(value);
  const formatSize = (size: number) => size >= 1048576 ? `${(size / 1048576).toFixed(1)} MB` : `${Math.max(1, Math.round(size / 1024))} KB`;

  function selected(document: Document) {
    file = document;
    password = '';
    showPassword = false;
    error = notice = savedPath = '';
    requestAnimationFrame(() => passwordInput?.focus());
  }

  async function choose() {
    if (busy) return;
    busy = true;
    error = '';
    try {
      const result = await desktop().SelectFile();
      if (result) selected(result);
    } catch (err) { error = message(err); }
    finally { busy = false; }
  }

  async function dropped(paths: string[]) {
    if (busy) return;
    busy = true;
    error = '';
    try { selected(await desktop().UseDroppedFile(paths)); }
    catch (err) { error = message(err); }
    finally { busy = false; }
  }

  async function clearFile() {
    if (busy) return;
    busy = true;
    try {
      await desktop().ClearFile();
      file = null;
      password = '';
      showPassword = false;
      error = notice = savedPath = '';
    } catch (err) { error = message(err); }
    finally { busy = false; }
  }

  async function unlock() {
    if (busy || !file || !password) return;
    busy = true;
    error = notice = savedPath = '';
    try {
      // Clear the visible input as soon as the bridge receives the password.
      const pending = desktop().RemovePassword(password);
      password = '';
      showPassword = false;
      const result = await pending;
      if (result.cancelled) notice = 'Save cancelled. No new file was written. Enter the password to try again.';
      else savedPath = result.path ?? '';
    } catch (err) { error = message(err); }
    finally { busy = false; }
  }

  onMount(() => {
    native = isDesktop();
    return onFileDrop((paths) => {
      if (page === 'password') void dropped(paths);
    });
  });
</script>

<div class="window-bar"><span>PDF Tools</span></div>
<div class="app-shell">
  <aside class="sidebar">
    <div class="brand"><div class="brand-icon"><Icon name="file" size={23} /></div><span>PDF Tools<span class="brand-subtitle">Your files. Your computer.</span></span></div>
    <div class="nav-label">YOUR WORKSPACE</div>
    <nav class="main-navigation" aria-label="Main navigation">
      <button class="nav-item" class:active={page === 'password'} aria-current={page === 'password' ? 'page' : undefined} onclick={() => navigate('password')} disabled={busy}>
        <Icon name="unlock" /><span>Remove Password</span>{#if page === 'password'}<span class="active-dot"></span>{/if}
      </button>
      <button class="nav-item" class:active={page === 'about'} aria-current={page === 'about' ? 'page' : undefined} onclick={() => navigate('about')} disabled={busy}>
        <Icon name="info" /><span>About</span>{#if page === 'about'}<span class="active-dot"></span>{/if}
      </button>
    </nav>
    <div class="sidebar-bottom"><span class="version">{appInfo.name} · v{appInfo.version}</span></div>
  </aside>

  <main>
    <header class="workspace-header"><span>PDF TOOLKIT <span class="breadcrumb-separator">/</span> <span class="breadcrumb-current">{page === 'about' ? 'About' : 'Remove Password'}</span></span><span class="offline-pill"><Icon name="shield" size={14} /> Offline & Private</span></header>
    {#if page === 'about'}
      <About />
    {/if}
    <div class="workspace" hidden={page !== 'password'}>
      <div class="page-heading"><h1>Unlock your PDF.</h1><p>Remove a password you already know.<br class="small-break" /> Keep a copy that’s easier to use.</p></div>
      {#if !native}<div class="preview-note">Interface preview · File processing is available in the desktop app.</div>{/if}

      <section class="tool-card" aria-label="Remove PDF password">
        <div class="card-heading"><span class="step-number">1</span><div><h2>Choose your document</h2><p>One PDF at a time, up to 100 MB.</p></div></div>
        {#if file}
          <div class="selected-file" style="--wails-drop-target: drop;"><div class="pdf-icon"><Icon name="file" size={27} /><span>PDF</span></div><div class="file-details"><strong title={file.path}>{file.name}</strong><span>{formatSize(file.size)} · Ready to unlock</span></div><button class="icon-button remove-file" onclick={clearFile} disabled={busy} aria-label="Remove selected file" title="Remove file"><Icon name="close" size={18} /></button></div>
          <button class="text-button change-file" onclick={choose} disabled={busy}>Choose a different PDF</button>
        {:else}
          <button class="drop-zone" onclick={choose} disabled={busy} style="--wails-drop-target: drop;" aria-label="Choose a PDF or drop a PDF here"><span class="upload-icon"><Icon name="upload" size={26} /></span><strong>Drop your PDF here</strong><span>or <span class="browse-link">browse files</span> on your computer</span><span class="file-type">PDF files only</span></button>
        {/if}

        <div class="section-divider"></div>
        <form onsubmit={(event) => { event.preventDefault(); void unlock(); }}>
          <div class="card-heading"><span class="step-number">2</span><div><h2>Enter the PDF password</h2><p>The password you use to open this document.</p></div></div>
          <label class="sr-only" for="password">PDF password</label>
          <div class="password-field"><Icon name="lock" size={18} /><input bind:this={passwordInput} id="password" type={showPassword ? 'text' : 'password'} bind:value={password} placeholder="Enter known password" disabled={!file || busy} autocomplete="off" autocapitalize="none" spellcheck="false" /><button type="button" class="icon-button" disabled={!file || busy} onclick={() => showPassword = !showPassword} aria-label={showPassword ? 'Hide password' : 'Show password'} aria-pressed={showPassword} title={showPassword ? 'Hide password' : 'Show password'}><Icon name="eye" size={19} /></button></div>
          <p class="password-note">Your password is used once and never saved.</p>
          {#if error}<div class="feedback error" role="alert">{error}</div>{/if}
          {#if notice}<div class="feedback notice" role="status">{notice}</div>{/if}
          {#if savedPath}<div class="feedback success" role="status"><Icon name="check" size={20} /><div><strong>Your unlocked PDF is saved.</strong><span class="saved-path">{savedPath}</span><span>The original PDF is unchanged.</span></div></div>{/if}
          <button class="primary-button" type="submit" disabled={!file || !password || busy}>{#if busy}<span class="spinner"></span><span>Working…</span>{:else}<Icon name="unlock" size={18} /><span>Remove password & save</span><Icon name="arrow" size={18} />{/if}</button>
          <div class="save-note" aria-live="polite"><Icon name="folder" size={14} /><span>{busy ? 'Unlocking your PDF, then choosing where to save.' : 'You’ll choose where to save your unlocked copy.'}</span></div>
        </form>
      </section>
    </div>
  </main>
</div>
