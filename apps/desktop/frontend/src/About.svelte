<script lang="ts">
  import Icon from './Icon.svelte';
  import { appInfo } from './app-info';
</script>

<div class="workspace about-page">
  <div class="page-heading">
    <h1>About {appInfo.name}</h1>
    <p>PDF tools that operate on your computer.</p>
  </div>

  <section class="tool-card about-card" aria-labelledby="app-details-heading">
    <div class="about-identity">
      <div class="brand-icon about-app-icon"><Icon name="file" size={28} /></div>
      <div>
        <h2 id="app-details-heading">{appInfo.name}</h2>
        <p>Version {appInfo.version}</p>
      </div>
    </div>

    <p class="about-description">
      Work with PDF files on your computer. Use tools to change your documents
      and save the results in a location that you select. Your files stay private.
    </p>

    <dl class="app-details">
      <div><dt>Version</dt><dd>{appInfo.version}</dd></div>
      <div><dt>Developer</dt><dd>{appInfo.developer}</dd></div>
      <div><dt>License</dt><dd>{appInfo.license}</dd></div>
    </dl>
  </section>

  <section class="tool-card about-privacy" aria-labelledby="privacy-heading">
    <div class="card-heading">
      <span class="privacy-icon"><Icon name="shield" size={23} /></span>
      <div>
        <h2 id="privacy-heading">Your files stay on your computer</h2>
        <p>The app does not send your files to a server.</p>
      </div>
    </div>
    <ul>
      <li>PDF processing operates without an internet connection.</li>
      <li>Your password is used for the operation. It is not saved.</li>
      <li>You do not need an account.</li>
    </ul>
  </section>
</div>

<style>
  .about-identity {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .about-app-icon {
    width: 52px;
    height: 56px;
    border-radius: 13px;
  }

  .about-identity h2 {
    font-size: 20px;
    letter-spacing: -.5px;
  }

  .about-identity p {
    color: #7a868d;
    font-size: 12px;
    margin: 4px 0 0;
  }

  .about-description {
    font-size: 13px;
    line-height: 1.8;
    color: #647078;
    margin: 22px 0;
  }

  .app-details {
    margin: 0;
  }

  .app-details > div {
    display: grid;
    grid-template-columns: minmax(90px, 1fr) 2fr;
    gap: 16px;
    padding: 15px 0;
    border-top: 1px solid #edf0f2;
    font-size: 12px;
    line-height: 1.6;
  }

  .app-details > div:last-child {
    padding-bottom: 0;
  }

  dt {
    color: #7a868d;
  }

  dd {
    margin: 0;
    overflow-wrap: anywhere;
    font-weight: 500;
  }

  .about-privacy .card-heading {
    align-items: center;
  }

  .about-privacy ul {
    margin: 18px 0 0;
    padding-left: 19px;
    color: #647078;
    font-size: 12px;
    line-height: 1.8;
  }

  .about-privacy li + li {
    margin-top: 7px;
  }
</style>
