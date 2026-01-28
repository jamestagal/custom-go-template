// Auto-generated bundle for pages/_index
// Unique components: whyChoose2425
// Imports common chunk with: footer, head, header, hero2436, html, nav, services2437

import common from '../../common.5fe3ef5b.js';

const unique = {
  'whyChoose2425': (props) => `<!-- ============================================ --><!--                 Why Choose                   --><!-- ============================================ --><section id="why-2425">
  <div class="cs-container">
    <div class="cs-content cs-title-group">
      <span class="cs-topper">${props.topper}</span>
      <h2 class="cs-title">${props.title}</h2>
      <picture class="cs-background cs-title-bg">
        <source media="(max-width: 767px)" srcset="${props.titleBackground.image_m.src}"></source>
        <source media="(min-width: 768px)" srcset="${props.titleBackground.image.src}"></source> 
        <img src="${props.titleBackground.image.src}" alt="${props.titleBackground.image.alt}" decoding="async"></img>
      </picture>
    </div>
    <div class="cs-content cs-cta">
      <h3 class="cs-h3">${props.cta.title}</h3>
      <a href="${props.cta.button.url}" class="cs-button-solid">${props.cta.button.text}</a> 
      <picture class="cs-background cs-cta-bg">
        <source media="(max-width: 767px)" srcset="${props.cta.background.image_m.src}"></source>
        <source media="(min-width: 768px)" srcset="${props.cta.background.image.src}"></source> 
        <img src="${props.cta.background.image.src}" alt="${props.cta.background.image.alt}" decoding="async"></img>
      </picture>
    </div>
    <div class="cs-text-group"><template x-for="text in props.textGroup">
      <p class="cs-text">
        <span x-text="text"></span>
      </p></template>
    </div>
    <ul class="cs-card-group"><template x-for="card in props.cards">
      <li class="cs-item">
        <div class="cs-wrapper" aria-hidden="true">
          <img class="cs-icon" loading="lazy" decoding="async" :src="card.icon.src" :alt="card.icon.alt"></img>
          <div class="cs-border cs-border1" aria-hidden="true"></div>
          <div class="cs-border cs-border2" aria-hidden="true"></div>
        </div>
        <div class="cs-info">
          <h3 class="cs-h3"><span x-text="card.title"></span></h3>
          <p class="cs-text">
            <span x-text="card.description"></span>
          </p>
        </div>
      </li></template>
    </ul>
    <!--Download svg and change fill color to match your clients brand color-->
    <img class="cs-graphic" src="${props.graphic.src}" alt="${props.graphic.alt}" decoding="async"></img>
  </div> 
</section><style>
    /*-- -------------------------- -->
<---         Services           -->
<--- -------------------------- -*/

/* Mobile - 360px */
@media only screen and (min-width: 0rem) {
  #why-2425 {
    padding: var(--sectionPadding);
    padding-bottom: 0;
    padding-top: 0;
    background-color: #030711;
    overflow: hidden;
    position: relative;
  }
  #why-2425 .cs-container {
    width: 100%;
    /* changes to 1280px at tablet */
    max-width: 36.5rem;
    margin: auto;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    gap: clamp(2rem, 9vw, 4rem);
    position: relative;
  }
  #why-2425 .cs-content {
    /* cs-content defines common rules of .cs-title-group and .cs-cta */
    text-align: left;
    width: 100%;
    padding: var(--sectionPadding);
    padding-left: 0;
    padding-right: 0;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;
    position: relative;
    z-index: 1;
  }
  #why-2425 .cs-cta {
    /* these rules override .cs-content where needed */
    text-align: center;
    align-items: center;
    padding: var(--sectionPadding);
    padding-left: 0;
    padding-right: 0;
    order: 9;
  }
  #why-2425 .cs-topper {
    color: var(--secondary);
  }
  #why-2425 .cs-title {
    color: var(--bodyTextColorWhite);
    width: 100%;
    max-width: 25.625rem;
    margin: 0;
  }
  #why-2425 .cs-button-solid {
    font-size: 1rem;
    line-height: clamp(2.875rem, 5.5vw, 3.5rem);
    text-decoration: none;
    font-weight: 700;
    text-align: center;
    color: #fff;
    white-space: nowrap;
    margin: 0;
    min-width: 11.875rem;
    padding: 0 1.6875rem;
    background-color: var(--primary);
    overflow: hidden;
    position: relative;
    z-index: 1;
  }
  #why-2425 .cs-button-solid:before {
    content: '';
    width: 100%;
    height: 100%;
    background: var(--secondary);
    position: absolute;
    top: 0;
    left: 0;
    z-index: -1;
    transform: scaleX(0);
    transform-origin: 0 0;
    transition: transform 0.3s ease;
    opacity: 1;
  }
  #why-2425 .cs-button-solid:hover:before {
    transform: scaleX(1);
  }
  #why-2425 .cs-text-group {
    display: flex;
    flex-direction: column;
    gap: clamp(1rem, 3.9vw, 1.5rem);
  }
  #why-2425 .cs-card-group {
    width: 100%;
    max-width: 90%;
    margin: 0 auto;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    gap: clamp(2rem, 7vw, 5rem) 0;
  }
  #why-2425 .cs-item {
    text-align: left;
    list-style: none;
    margin: 0;
    display: flex;
    justify-content: flex-start;
    align-items: flex-start;
    grid-column: span 12;
    gap: clamp(1.5rem, 3.1vw, 2rem);
    position: relative;
    z-index: 1;
  }
  #why-2425 .cs-wrapper {
    width: 4rem;
    height: 4rem;
    display: block;
    flex: none;
    position: relative;
    z-index: 1;
  }
  #why-2425 .cs-border {
    width: 100%;
    height: 100%;
    position: absolute;
    top: 0;
    left: 0;
    z-index: -1;
  }
  #why-2425 .cs-border:before {
    content: '';
    opacity: 1;
    display: block;
    position: absolute;
    top: 0;
    left: 0;
    z-index: -1;
  }
  #why-2425 .cs-border:after {
    content: '';
    opacity: 1;
    display: block;
    position: absolute;
    top: 0;
    left: 0;
    z-index: -1;
  }
  #why-2425 .cs-border1 {
    /* top left border */
  }
  #why-2425 .cs-border1:before {
    width: 100%;
    height: 1px;
    background: linear-gradient(to right, #88898b 0%, rgba(0, 0, 0, 0) 78%);
  }
  #why-2425 .cs-border1:after {
    width: 1px;
    height: 100%;
    background: linear-gradient(to bottom, #88898b 0%, rgba(0, 0, 0, 0) 78%);
  }
  #why-2425 .cs-border2:before {
    width: 100%;
    height: 1px;
    background: linear-gradient(to right, rgba(0, 0, 0, 0) 22%, #88898b 100%);
    top: auto;
    bottom: 0;
  }
  #why-2425 .cs-border2:after {
    width: 1px;
    height: 100%;
    background: linear-gradient(to bottom, rgba(0, 0, 0, 0) 22%, #88898b 100%);
    left: auto;
    right: 0;
  }
  #why-2425 .cs-icon {
    width: 1.5rem;
    height: 1.5rem;
    margin: auto;
    display: block;
    position: absolute;
    inset: 0;
    z-index: 2;
  }
  #why-2425 .cs-info {
    font-size: clamp(0.875rem, 1.5vw, 1rem);
    line-height: 1.5em;
    margin: 0;
    padding: 0;
  }
  #why-2425 .cs-h3 {
    font-size: clamp(1.25rem, 2.5vw, 1.5625rem);
    font-weight: 500;
    line-height: 1.2em;
    text-align: inherit;
    max-width: 24ch;
    margin: 0 0 clamp(0.75rem, 1.8vw, 1rem) 0;
    color: var(--bodyTextColorWhite);
  }
  #why-2425 .cs-text {
    color: var(--bodyTextColorWhite);
    opacity: 0.9;
  }
  #why-2425 .cs-background {
    /* common decorative background image rules defined here */
    /* specific rules defined on .cs-title-bg and .cs-cta below */
    height: 100%;
    display: block;
    position: absolute;
    top: 0;
    left: 0;
    overflow: hidden;
    z-index: -1;
  }
  #why-2425 .cs-background:before {
    /* common pseudo-element rules defined here */
    /* gradient overlays defined on .cs-title-bg and .cs-cta below */
    content: '';
    width: 100%;
    height: 100%;
    display: block;
    position: absolute;
    bottom: -1px;
    right: 0;
    z-index: -1;
  }
  #why-2425 .cs-background img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    position: relative;
    z-index: -2;
  }
  #why-2425 .cs-title-bg {
    width: 100vw;
    left: 50%;
    transform: translateX(-50%);
  }
  #why-2425 .cs-title-bg:before {
    height: 80%;
    background: linear-gradient(to bottom, rgba(0, 0, 0, 0) 0%, #03060f 90%, #030711 100%);
  }
  #why-2425 .cs-cta-bg {
    width: 100vw;
    left: 50%;
    transform: translateX(-50%);
  }
  #why-2425 .cs-cta-bg:before {
    /* horizontal gradient */
    background: linear-gradient(to right, #030711 0%, #03060f 10%, rgba(0, 0, 0, 0) 100%);
  }
  #why-2425 .cs-graphic {
    width: auto;
    height: 32.0625rem;
    position: absolute;
    left: calc(50% + 50vw);
    bottom: 36%;
    transform: translate(-100%, 50%);
    display: none;
  }
}
/* Tablet - 768px */
@media only screen and (min-width: 48rem) {
  #why-2425 {
    /* added padding back to section container for a consistent padding bottom */
    padding: var(--sectionPadding);
    padding-top: 0;
    flex-direction: row;
    flex-wrap: wrap;
  }
  #why-2425 .cs-container {
    max-width: 80rem;
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    align-items: stretch;
    gap: 4rem 0;
  }
  #why-2425 .cs-content {
    grid-column: span 6;
    justify-content: flex-end;
  }
  #why-2425 .cs-title-group {
    padding-top: clamp(8.75rem, 20vw, 15rem);
    padding-bottom: 3.75rem;
    justify-content: flex-end;
  }
  #why-2425 .cs-topper {
    max-width: none;
    margin: 0;
  }
  #why-2425 .cs-title {
    max-width: none;
    margin: 0;
  }
  #why-2425 .cs-cta {
    text-align: left;
    padding-top: 6.25rem;
    padding-left: clamp(2rem, 4vw, 3.75rem);
    padding-bottom: 3.75rem;
    align-items: flex-start;
    order: unset;
  }
  #why-2425 .cs-text-group {
    width: 100%;
    flex-direction: row;
    gap: 1.5rem;
    grid-column: span 12;
  }
  #why-2425 .cs-text-group .cs-text {
    max-width: none;
  }
  #why-2425 .cs-card-group {
    margin: 0 auto 0 0;
    gap: 3rem 0;
    grid-column: span 12;
  }
  #why-2425 .cs-item {
    padding-inline: 0 2rem;
    flex-direction: column;
    grid-column: span 6;
  }
  #why-2425 .cs-graphic {
    display: block;
  }
  #why-2425 .cs-title-bg {
    width: 50vw;
    left: auto;
    right: 0;
    transform: none;
  }
  #why-2425 .cs-cta-bg {
    width: 50vw;
    margin: 0;
    left: 0;
    transform: none;
  }
}
/* Large Desktop - 1300px */
@media only screen and (min-width: 81.25rem) {
  #why-2425 .cs-title-group {
    padding-bottom: 0;
    grid-column: span 8;
  }
  #why-2425 .cs-title-bg {
    width: 65vw;
  }
  #why-2425 .cs-cta {
    grid-column: span 4;
  }
  #why-2425 .cs-cta-bg {
    width: 45vw;
  }
  #why-2425 .cs-text-group {
    margin-bottom: 4rem;
  }
  #why-2425 .cs-card-group {
    max-width: 100%;
  }
  #why-2425 .cs-item {
    grid-column: span 4;
  }
}
</style>`
};

export default { ...common, ...unique };
