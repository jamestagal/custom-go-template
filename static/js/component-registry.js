export default {
  '../components/footer-old.html': (props) => `<footer class="footer">
  <div class="footer-container">
    <div class="footer-grid">
      <div>
        <div class="footer-brand">${props.companyName}</div>
        <p class="footer-description">
          Our custom template engine makes building reactive web applications simple and efficient.
          We focus on developer experience without sacrificing performance.
        </p>
        <div class="social-links"><template x-for="social,  in socialLinks">
            <a href="${props.social.url}" class="social-link" title="${props.social.label}">
              ${props.social.icon }
            </a></template>
        </div>
      </div>
      
      <div>
        <h3 class="footer-heading">Quick Links</h3>
        <ul class="footer-links"><template x-for="link, index in links.slice(0, 3)">
            <li class="footer-link">
              <a href="${props.link.url}">${props.link.label}</a>
            </li></template>
        </ul>
      </div>
      
      <div>
        <h3 class="footer-heading">Resources</h3>
        <ul class="footer-links"><template x-for="link, index in links.slice(3)">
            <li class="footer-link">
              <a href="${props.link.url}">${props.link.label}</a>
            </li></template>
        </ul>
      </div>
    </div>
    
    <div class="copyright">
      &copy; ${props.year} ${props.companyName}. All rights reserved.
    </div>
  </div>
</footer>`,
  'Hero2436': (props) => `<!-- ============================================ --><!--                   Hero                       --><!-- ============================================ --><section id="hero-2436">
    <div class="cs-container">
        <div class="cs-content">
            <div class="cs-flex cs-flex1">
                <span class="cs-topper">${props.topper}</span>
                <h2 class="cs-title">${props.title}</h2>
            </div>
            <div class="cs-flex cs-flex2">
                <p class="cs-text">
                    ${props.description}
                </p>
                <a href="${props.buttonLink}" class="cs-button-solid">${props.buttonText}</a>
            </div>
        </div>
        <img class="cs-logo-feature" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/Logos/art-logo.svg" alt="logo" width="1280" height="239"></img>
    </div>
    <!--Background Image-->
    <picture class="cs-background">
        <!--Mobile Image-->
        <source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art4.jpg"></source>
        <!--Tablet and above Image-->
        <source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art4.jpg"></source>
        <img decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art4.jpg" alt="face with body art" width="1800" height="860" aria-hidden="true" fetchpriority="high"></img>
    </picture>
</section>`,
  'Footer-old': (props) => `<footer class="footer">
  <div class="footer-container">
    <div class="footer-grid">
      <div>
        <div class="footer-brand">${props.companyName}</div>
        <p class="footer-description">
          Our custom template engine makes building reactive web applications simple and efficient.
          We focus on developer experience without sacrificing performance.
        </p>
        <div class="social-links"><template x-for="social,  in socialLinks">
            <a href="${props.social.url}" class="social-link" title="${props.social.label}">
              ${props.social.icon }
            </a></template>
        </div>
      </div>
      
      <div>
        <h3 class="footer-heading">Quick Links</h3>
        <ul class="footer-links"><template x-for="link, index in links.slice(0, 3)">
            <li class="footer-link">
              <a href="${props.link.url}">${props.link.label}</a>
            </li></template>
        </ul>
      </div>
      
      <div>
        <h3 class="footer-heading">Resources</h3>
        <ul class="footer-links"><template x-for="link, index in links.slice(3)">
            <li class="footer-link">
              <a href="${props.link.url}">${props.link.label}</a>
            </li></template>
        </ul>
      </div>
    </div>
    
    <div class="copyright">
      &copy; ${props.year} ${props.companyName}. All rights reserved.
    </div>
  </div>
</footer>`,
  'Headerlogo': (props) => `<header class="header">
  <div class="header-container">
    <a href="/" class="brand">
      <svg width="36.206mm" height="7.781mm" version="1.1" viewBox="0 0 36.206 7.781" xmlns="http://www.w3.org/2000/svg">
        <g transform="translate(-264.95 11.104)">
          <path d="m273.59-6.453-2e-3 1.6107c-7e-3 0.22466-0.12785 0.35734-0.35253 0.36241-0.25549 0-0.36051-0.14631-0.36051-0.36123v-4.6353c0-0.092936 0.0232-0.16264 0.0696-0.20911 0.0464-0.052254 0.11906-0.092936 0.2178-0.12198 0.16852-0.052248 0.37175-0.090037 0.6099-0.11327 0.24404-0.023229 0.47924-0.034889 0.70576-0.034889 0.7377 0 1.2751 0.15683 1.612 0.4705 0.34263 0.30786 0.51403 0.72609 0.51403 1.2547 0 0.54021-0.17425 0.97296-0.52275 1.2982-0.34852 0.31948-0.88589 0.47922-1.612 0.47922zm0.83646-0.59249c0.46468 0 0.82484-0.095841 1.0804-0.28753 0.25564-0.19168 0.38343-0.48792 0.38343-0.88872 0-0.39499-0.12489-0.68252-0.37466-0.86259-0.24395-0.18588-0.59546-0.27882-1.0543-0.27882-0.1568 0-0.31363 0.00874-0.47042 0.026165-0.15106 0.011654-0.2876 0.02902-0.40959 0.052251v2.2393z">
          <path d="m290.02-4.9895c0.13651 0 0.23267-0.00934 0.28845-0.027942 0.062-0.024852 0.12724-0.037205 0.19548-0.037205 0.068 0 0.13031 0.027942 0.18614 0.083763 0.0561 0.049626 0.0837 0.12719 0.0837 0.23267 0 0.099268-0.0499 0.17062-0.1489 0.21405-0.21711 0.099268-0.40323 0.14891-0.55839 0.14891-0.15512 0-0.29778-0.00934-0.42812-0.027942-0.12403-0.012426-0.24506-0.052715-0.36289-0.12099-0.27922-0.16132-0.41883-0.44672-0.41883-0.8562v-2.3639h-0.53971c-0.0931 0-0.24964-0.10621-0.25249-0.29294-2e-3 -0.17599 0.15385-0.26609 0.23168-0.26866l0.55806-0.00462 0.0103-0.64389c5e-3 -0.12544 0.16669-0.25549 0.33632-0.25619 0.18493 7.719e-4 0.34881 0.12909 0.3458 0.26617v0.60493h0.86549c0.0871 0 0.15818 0.027942 0.21402 0.083754 0.056 0.055807 0.0837 0.12719 0.0837 0.21405 0 0.086862-0.0279 0.15821-0.0837 0.21405-0.0562 0.055807-0.12713 0.083761-0.21402 0.083761h-0.86549v2.308c0 0.19234 0.0588 0.32263 0.17683 0.39087 0.062 0.037205 0.16125 0.055884 0.29781 0.055884z">
          <g transform="matrix(.77185 0 0 .77185 312.99 -65.312)" fill-rule="evenodd" stroke-width=".26458px">
            <path d="m-54.57 74.523c-1.7919 1.907-5.3906 2.3551-4.3001 4.9881 0.36199 0.74862-0.27419 1.0155-0.50107 0.61951 0 0-3.4086-2.8786-1.4504-4.6156 0.98084-1.0161 6.9801-5.3003 6.3041-1.3029z" fill="#1c7fc7">
            <path d="m-57.629 73.238s-1.5075 0.83503-2.7039 1.8531c0 0 3.0859-1.3869 5.7631-0.56799 0 0 0.39145-1.6984-3.0593-1.2851z" fill="#004a77" fill-opacity=".87434">
            <path d="m-62.252 72.183s0.38056-1.1878 2.1798-1.9508l0.96418 0.28985s6.7226-1.3794 4.5374 4.0016c0 0 0.5771-2.071-4.5695-1.033l-0.70199 0.43123c-1.906-0.26683-2.4099-1.7389-2.4099-1.7389" fill="#22a6ed">
          </g>
          <path d="m277.62-9.6188q0-0.14572 0.10242-0.2368 0.10234-0.10019 0.25131-0.10019 0.14885 0 0.24197 0.10019 0.10234 0.091077 0.10234 0.2368v4.8453q0 0.13662-0.10234 0.2368-0.0931 0.10019-0.24197 0.10019-0.14898 0-0.25131-0.10019-0.10242-0.10018-0.10242-0.2368z">
          <path d="m280.03-6.1675q0.0651 0.51186 0.39083 0.81897 0.33506 0.29781 0.91202 0.29781 0.58636 0 0.99581-0.18613 0.13033-0.065145 0.23266-0.065145 0.10242-0.00934 0.18612 0.074482 0.0931 0.083761 0.0931 0.18613 0 0.20474-0.16752 0.2885-0.15823 0.083758-0.28845 0.14891-0.13034 0.065144-0.27927 0.11168-0.34429 0.093063-0.80033 0.093063-0.93997 0-1.4611-0.52116-0.51184-0.53048-0.51184-1.4984 0-0.83759 0.43737-1.396 0.49331-0.62354 1.3866-0.62354 0.85624 0 1.3495 0.577 0.46533 0.53978 0.46533 1.3494 0 0.1396-0.10234 0.24197-0.0931 0.10237-0.24196 0.10237zm1.126-1.6845q-0.67936 0-0.9865 0.60492-0.11163 0.21405-0.13952 0.51186h2.2521q-0.0277-0.53978-0.4002-0.85621-0.2978-0.26058-0.72589-0.26058z">
          <path d="m285.85-7.7962q-0.67009 0-1.2564 0.73522v2.2801q0 0.1396-0.10244 0.24197-0.10232 0.10237-0.24199 0.10237-0.13951 0-0.24199-0.10237-0.10232-0.10237-0.10232-0.24197v-3.2666q0-0.14891 0.093-0.25128 0.10246-0.10237 0.2513-0.10237 0.13964 0 0.24199 0.10237 0.10244 0.093062 0.10244 0.25128v0.31642q0.30708-0.31642 0.53042-0.45602 0.42814-0.25128 0.81904-0.25128 0.40014 0 0.65139 0.12099 0.25126 0.11167 0.42812 0.32573 0.35368 0.41879 0.35368 1.0423v2.1684q0 0.1396-0.10242 0.24197-0.10234 0.10237-0.24195 0.10237-0.13958 0-0.24199-0.10237-0.10234-0.10237-0.10234-0.24197v-2.094q0-0.42811-0.19544-0.67007-0.19549-0.25127-0.64217-0.25127z">
          <path d="m291.49-8.0475q0-0.15821 0.10242-0.25128 0.10234-0.10237 0.25127-0.10237 0.14887 0 0.24197 0.10237 0.093 0.093062 0.093 0.25128v3.2666q0 0.1396-0.093 0.24197-0.0931 0.10237-0.24197 0.10237-0.14894 0-0.25127-0.10237-0.10242-0.10237-0.10242-0.24197zm0.76312-1.4983q0 0.16751-0.11161 0.27919-0.11174 0.11168-0.27921 0.11168h-0.0465q-0.16751 0-0.27921-0.11168-0.10234-0.11168-0.10234-0.27919v-0.027942q0-0.16752 0.11165-0.2792 0.11167-0.11167 0.26989-0.11167h0.0465q0.16747 0 0.27921 0.11167 0.11161 0.11168 0.11161 0.2792z">
          <path d="m295.26-5.0508q0.42813 0 0.80971-0.20474 0.13025-0.074483 0.24196-0.074483 0.11166-0.00934 0.19538 0.093062 0.0839 0.10237 0.0839 0.24197 0 0.1303-0.17686 0.24198-0.53977 0.35365-1.2098 0.35365-0.84694 0-1.4332-0.53047-0.61426-0.5677-0.60495-1.489 0-0.92135 0.60495-1.489 0.57697-0.53978 1.4332-0.53047 0.45601 0 0.74448 0.12099 0.29781 0.11167 0.46533 0.23266 0.17686 0.11168 0.17686 0.25128 0 0.1396-0.0839 0.23267-0.0837 0.093062-0.16746 0.093062-0.13033 0-0.26988-0.074483-0.39088-0.21405-0.80971-0.20474-0.67006 0-1.0423 0.38156-0.363 0.37226-0.363 0.98649 0 0.61423 0.363 0.99579 0.37224 0.37226 1.0423 0.37226z">
          <path d="m299.21-8.4383q0.88411 0 1.4239 0.55839 0.53971 0.5677 0.52112 1.4611 0 0.90273-0.52112 1.4611-0.53979 0.55839-1.4239 0.55839-0.89347 0-1.4146-0.55839-0.53979-0.55839-0.53979-1.4611 0-0.91204 0.53979-1.4611 0.52111-0.55839 1.4146-0.55839zm-0.85627 3.0991q0.18617 0.15821 0.40954 0.23267 0.23266 0.074482 0.44673 0.074482 0.21402 0 0.43739-0.074482 0.22337-0.074483 0.40943-0.24197 0.4095-0.37226 0.4095-1.0796 0-0.68868-0.4095-1.0702-0.34429-0.31642-0.84682-0.30712-0.82832 0-1.1447 0.74453-0.11171 0.26988-0.11171 0.64215 0 0.37226 0.11171 0.64215 0.11164 0.26989 0.28846 0.43741z">
        </g>
      </svg>
    </a>

    <nav class="nav">
      <a href="/" class="nav-item">Home</a>
      <a href="/products" class="nav-item">Products</a>
      <a href="/login" class="nav-item">Login</a>
      <a href="/register" class="nav-item">Register</a>
    </nav>
  </div>
</header>`,
  'Html': (props) => `<html lang="en">

<!-- Global Head Component -->

<body>
	<!-- User Menu (conditional - if authenticated) -->
	<!-- {#if user && $user.isAuthenticated}
		<UserMenu user={user} content={content} shadowContent={shadowContent} />
	{/if} -->

	<!-- Login Modal (conditional) -->
	<!-- <LoginModal /> -->

	<main>
		<!-- Global Navigation -->

		<!-- Global Header -->
		<!-- <Header /> -->

		<!-- ============================================ -->
		<!-- DYNAMIC CONTENT LAYOUT INJECTION             -->
		<!-- This is where page-specific layouts render  -->
		<!-- For pages â renders layouts/content/pages.html -->
		<!-- For courses â renders layouts/content/courses.html -->
		<!-- ============================================ -->

		<!-- Dynamic layout component injection -->
		<!-- In Svelte this is: <svelte:component this={layout} {...content.fields} /> -->
		<!-- In our Go template engine, this is handled by the server -->
		<!-- The server will inject the appropriate layout based on the route -->

		<!-- Slot for dynamic content - replaced by server with specific layout -->
		<!-- CRITICAL FIX: Pass content and allContent explicitly to layout component -->
		<!-- Pages.html needs content.components for iteration -->
		<!-- Newsletter Subscription (optional global component) -->
		<!-- <Newsletter /> -->

		<!-- Global Footer -->
	</main>

</body>

</html>`,
  'Nav': (props) => `<!-- ============================================ --><!--                 Navigation                   --><!-- ============================================ --><header id="cs-navigation" :class="{'cs-active': mobileMenuOpen}">
    <div class="cs-container">
        <div class="cs-top">
            <!--Navigation List-->
            <nav class="cs-nav" aria-label="main navigation">
                <!--Mobile Nav Toggle-->
                <button class="cs-toggle" @click="toggleMobileMenu()" :aria-expanded="mobileMenuOpen" aria-label="mobile menu toggle" aria-controls="cs-ul-wrapper" aria-haspopup="menu">
                    <span class="cs-box" aria-hidden="true">
                        <span class="cs-line cs-line1"></span>
                        <span class="cs-line cs-line2"></span>
                        <span class="cs-line cs-line3"></span>
                    </span>
                </button>
                <!-- We need a wrapper div so we can set a fixed height on the cs-ul in case the nav list gets too long from too many dropdowns being opened and needs to have an overflow scroll. This wrapper acts as the background so it can go the full height of the screen and not cut off any overflowing nav items while the cs-ul stops short of the bottom of the screen, which keeps all nav items in view no matter how mnay there are-->
                <div class="cs-ul-wrapper" id="cs-ul-wrapper">
                    <ul class="cs-ul">
                        <li class="cs-li">
                            <a href="" class="cs-li-link cs-active" aria-current="page">
                                Home
                            </a>
                        </li>
                        <li class="cs-li">
                            <a href="" class="cs-li-link">
                                About
                            </a>
                        </li>
                        <!--Copy and paste this cs-dropdown list item and replace any .cs-li with this cs-dropdown group to make a new dropdown and it will work-->
                        <!-- Note: When copying this dropdown, update the id attribute with an appropriate name (e.g., "services-dropdown-toggle", "products-dropdown-toggle") and ensure the aria-labelledby attribute on the ul matches the button's id -->
                        <li class="cs-li cs-dropdown">
                            <button class="cs-li-link cs-dropdown-toggle" aria-expanded="false" aria-haspopup="menu" id="services-dropdown-toggle">
                                Services
                                <img class="cs-drop-icon" src="https://csimages2.nyc3.digitaloceanspaces.com/Icons/chevron-dropdown.svg" alt="" width="15" height="15" decoding="async" aria-hidden="true"></img>
                            </button>
                            <ul class="cs-drop-ul" aria-labelledby="services-dropdown-toggle">
                                <li class="cs-drop-li">
                                    <a href="" class="cs-li-link cs-drop-link">Service 1</a>
                                </li>
                                <li class="cs-drop-li">
                                    <a href="" class="cs-li-link cs-drop-link">Service 2</a>
                                </li>
                            </ul>
                        </li>
                    </ul>
                    <ul class="cs-ul">
                        <li class="cs-li cs-dropdown">
                            <button class="cs-li-link cs-dropdown-toggle" aria-expanded="false" aria-haspopup="menu" id="services-dropdown-toggle">
                                Artists
                                <img class="cs-drop-icon" src="https://csimages2.nyc3.digitaloceanspaces.com/Icons/chevron-dropdown.svg" alt="" width="15" height="15" decoding="async" aria-hidden="true"></img>
                            </button>
                            <ul class="cs-drop-ul" aria-labelledby="services-dropdown-toggle">
                                <li class="cs-drop-li">
                                    <a href="" class="cs-li-link cs-drop-link">Artist 1</a>
                                </li>
                                <li class="cs-drop-li">
                                    <a href="" class="cs-li-link cs-drop-link">Artist 2</a>
                                </li>
                            </ul>
                        </li>
                        <li class="cs-li">
                            <a href="" class="cs-li-link">FAQ</a>
                        </li>
                        <li class="cs-li">
                            <a href="" class="cs-li-link">Contact</a>
                        </li>
                    </ul>
                </div>
                <!--Dark Mode toggle, uncomment button code if you want to enable a dark mode toggle-->
                <button id="dark-mode-toggle" aria-label="dark mode toggl" aria-pressed="false">
                    <svg class="cs-moon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 480 480" style="enable-background:new 0 0 480 480" xml:space="preserve" aria-hidden="true">
                        <path d="M459.782 347.328c-4.288-5.28-11.488-7.232-17.824-4.96-17.76 6.368-37.024 9.632-57.312 9.632-97.056 0-176-78.976-176-176 0-58.4 28.832-112.768 77.12-145.472 5.472-3.712 8.096-10.4 6.624-16.832S285.638 2.4 279.078 1.44C271.59.352 264.134 0 256.646 0c-132.352 0-240 107.648-240 240s107.648 240 240 240c84 0 160.416-42.688 204.352-114.176 3.552-5.792 3.04-13.184-1.216-18.496z">
                    </svg>
                    <img class="cs-sun" aria-hidden="true" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Icons%2Fsun.svg" decoding="async" alt="" width="15" height="15"></img>
                </button>
            </nav>

        </div>
        <div class="cs-bottom">
            <!--Nav Logo-->
            <a href="" class="cs-logo" aria-label="Return to [Company Name]home page">
                <img src="https://csimages2.nyc3.digitaloceanspaces.com/Icons/artistitch-logo.svg" alt="" width="210" height="29" aria-hidden="true" decoding="async"></img>
            </a>
        </div>
    </div>
</header>`,
  '_index': (props) => ``,
  'Comprehensive': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>Comprehensive Template Test</title>
</head>
<body>

<style>
  .container 
    max-width: 1200px;
    margin: 0 auto;
    padding: 1rem;
  

  .section 
    margin-bottom: 2rem;
    border: 1px solid #eaeaea;
    border-radius: 0.5rem;
    padding: 1.5rem;
  

  .section-title 
    font-size: 1.5rem;
    font-weight: bold;
    margin-bottom: 1rem;
    color: #333;
  

  .card 
    border: 1px solid #ddd;
    border-radius: 0.25rem;
    padding: 1rem;
    margin-bottom: 1rem;
  

  .product-grid 
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1rem;
  

  .tag 
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    margin-right: 0.5rem;
  

  .notification 
    padding: 0.75rem;
    border-radius: 0.25rem;
    margin-bottom: 0.5rem;
  

  .notification-info 
    background-color: #e0f2fe;
    border-left: 4px solid #38bdf8;
  

  .notification-success 
    background-color: #dcfce7;
    border-left: 4px solid #4ade80;
  

  .notification-warning 
    background-color: #fef3c7;
    border-left: 4px solid #fbbf24;
  
</style>

<div class="container">
  <!-- Static Component Usage -->

  <!-- 1. Basic Expressions Section -->
  <div class="section">
    <h2 class="section-title">1. Basic Expressions</h2>
    <div class="card">
      <p>Welcome to ${props.title}!</p>
      <p>Current user: <strong>${props.user.name}</strong> (${props.user.role })</p>
      <p>${props.getGreeting() }, ${props.user.name }!</p>
      <p>Number of products: ${props.products.length }</p>
      <p>Number of categories: ${props.categories.length }</p>
      <p>Current theme: ${props.settings.theme }</p>
      <p>Currency: ${props.settings.currency }</p>
    </div>
  </div>

  <!-- 2. Conditionals Section -->
  <div class="section">
    <h2 class="section-title">2. Conditionals</h2>

    <!-- Simple if condition -->
    <div class="card">
      <h3>User Authentication Status</h3><template x-if="isLoggedIn">
        <p>You are currently logged in as ${props.user.name }.</p>
        <p>Email: ${props.user.email }</p></template><template x-else>
        <p>You are not logged in. Please sign in to access your account.</p></template>
    </div>

    <!-- if/else if/else condition -->
    <div class="card">
      <h3>User Role Check</h3><template x-if="user.role === "admin"">
        <p>Welcome, Administrator! You have full access to the system.</p></template><template x-else-if="user.role === "manager"">
        <p>Welcome, Manager! You can manage products and users.</p></template><template x-else-if="user.role === "editor"">
        <p>Welcome, Editor! You can edit product information.</p></template><template x-else>
        <p>Welcome, User! You have standard user privileges.</p></template>
    </div>

    <!-- Nested conditionals -->
    <div class="card">
      <h3>Product Availability</h3><template x-if="filteredProducts.length > 0">
        <p>We have ${props.filteredProducts.length } products matching your filters:</p><template x-if="settings.filters.inStockOnly">
          <p>Showing only in-stock items.</p></template><template x-else>
          <p>Showing all items (including out of stock).</p></template></template><template x-else>
        <p>No products match your current filters.</p></template>
    </div>
  </div>

  <!-- 3. Loops Section -->
  <div class="section">
    <h2 class="section-title">3. Loops</h2>

    <!-- Simple array loop -->
    <div class="card">
      <h3>Product List</h3>
      <ul><template x-for="product,  in filteredProducts">
          <li>${props.product.name} - ${props.formatPrice(product.price) }<template x-if="!product.inStock">
              <span style="color: red;"> (Out of Stock)</span></template>
          </li></template>
      </ul>
    </div>

    <!-- Loop with index -->
    <div class="card">
      <h3>Numbered Product List</h3>
      <ol><template x-for="product, index in filteredProducts">
          <li value="${index + 1 }">
            ${props.product.name} - ${props.formatPrice(product.price)}<template x-if="product.featured">
              <span style="color: gold;"> ★ Featured</span></template>
          </li></template>
      </ol>
    </div>

    <!-- Object loop -->
    <div class="card">
      <h3>Settings Configuration</h3>
      <dl><template x-for="key, value in settings">
          <dt>${props.key }:</dt><template x-if="typeof value === 'object'">
            <dd>
              <dl><template x-for="subKey, subValue in value">
                  <dt>${props.subKey}:</dt>
                  <dd>${props.subValue}</dd></template>
              </dl>
            </dd></template><template x-else>
            <dd>${props.value}</dd></template></template>
      </dl>
    </div>

    <!-- Nested loops -->
    <div class="card">
      <h3>Categories and Products</h3><template x-for="category,  in categories">
        <div style="margin-bottom: 1rem;">
          <h4>${props.category.name} (${props.category.items.length} items)</h4><template x-if="category.items.length > 0">
            <ul><template x-for="item,  in category.items">
                <li>
                  ${item.name} - ${props.formatPrice(item.price)}
                  <div>
                    Tags:
                    <template x-for="tag,  in item.tags">
                      <span class="tag ${props.getTagClass(tag) }">${props.tag}</span></template>
                  </div>
                </li></template>
            </ul></template><template x-else>
            <p>No items in this category.</p></template>
        </div></template>
    </div>
  </div>

  <!-- 4. Components Section -->
  <div class="section">
    <h2 class="section-title">4. Components</h2>

    <!-- Static components with props -->
    <div class="card">
      <h3>User Profile Component</h3>
    </div>

    <!-- Component loop -->
    <div class="card">
      <h3>Product Cards</h3>
      <div class="product-grid"><template x-for="product,  in filteredProducts.filter(p => p.featured)"></template>
      </div>
    </div>

    <!-- Conditional components -->
    <div class="card">
      <h3>Notifications</h3><template x-for="notification,  in notifications"></template>
    </div>

    <!-- Dynamic components -->
    <div class="card">
      <h3>Dynamic Component Example</h3><template x-if="user.role === "admin""></template><template x-else></template>
    </div>
  </div>

  <!-- 5. Advanced Features Section -->
  <div class="section">
    <h2 class="section-title">5. Advanced Features</h2>

    <!-- Computed values and functions -->
    <div class="card">
      <h3>Filtered Products (Computed)</h3>
      <p>Products between ${props.settings.filters.minPrice } and ${props.settings.filters.maxPrice}:</p>
      <ul><template x-for="product,  in filteredProducts">
          <li>${props.product.name} - ${props.formatPrice(product.price)}</li></template>
      </ul>
    </div>

    <!-- Complex expressions -->
    <div class="card">
      <h3>Complex Expressions</h3>
      <p>Average product price: ${props.formatPrice(products.reduce((props.sum, p) => props.sum + p.price, 0) / props.products.length)}</p>
      <p>Featured products: ${props.products.filter(p => p.featured).length}</p>
      <p>In-stock products: ${props.products.filter(p => p.inStock).length}</p>
      <p>Total inventory value: ${props.formatPrice(products.filter(p => p.inStock).reduce((sum, p) => sum + p.price, 0))}</p>
    </div>

    <!-- Combining features -->
    <div class="card">
      <h3>Combined Features Example</h3><template x-if="filteredProducts.length > 0">
        <div>
          <p>Top ${Math.min(3, filteredProducts.length)} products:</p>
          <div class="product-grid"><template x-for="product, index in filteredProducts.slice(0, 3)">
              <div class="card">
                <h4>${index + 1}. ${props.product.name}</h4>
                <p>Price: ${props.formatPrice(product.price)}</p>
                <div><template x-if="product.inStock">
                    <span style="color: green;">In Stock</span></template><template x-else>
                    <span style="color: red;">Out of Stock</span></template>
                </div>
                <div>
                  Tags:
                  <template x-for="tag,  in product.tags">
                    <span class="tag ${props.getTagClass(tag) }">${props.tag}</span></template>
                </div>
              </div></template>
          </div>
        </div></template><template x-else>
        <p>No products available.</p></template>
    </div>
  </div>

  <!-- Footer component -->
</div>

</body>
</html>`,
  '../content/comprehensive.html': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>Comprehensive Template Test</title>
</head>
<body>

<style>
  .container 
    max-width: 1200px;
    margin: 0 auto;
    padding: 1rem;
  

  .section 
    margin-bottom: 2rem;
    border: 1px solid #eaeaea;
    border-radius: 0.5rem;
    padding: 1.5rem;
  

  .section-title 
    font-size: 1.5rem;
    font-weight: bold;
    margin-bottom: 1rem;
    color: #333;
  

  .card 
    border: 1px solid #ddd;
    border-radius: 0.25rem;
    padding: 1rem;
    margin-bottom: 1rem;
  

  .product-grid 
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1rem;
  

  .tag 
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    margin-right: 0.5rem;
  

  .notification 
    padding: 0.75rem;
    border-radius: 0.25rem;
    margin-bottom: 0.5rem;
  

  .notification-info 
    background-color: #e0f2fe;
    border-left: 4px solid #38bdf8;
  

  .notification-success 
    background-color: #dcfce7;
    border-left: 4px solid #4ade80;
  

  .notification-warning 
    background-color: #fef3c7;
    border-left: 4px solid #fbbf24;
  
</style>

<div class="container">
  <!-- Static Component Usage -->

  <!-- 1. Basic Expressions Section -->
  <div class="section">
    <h2 class="section-title">1. Basic Expressions</h2>
    <div class="card">
      <p>Welcome to ${props.title}!</p>
      <p>Current user: <strong>${props.user.name}</strong> (${props.user.role })</p>
      <p>${props.getGreeting() }, ${props.user.name }!</p>
      <p>Number of products: ${props.products.length }</p>
      <p>Number of categories: ${props.categories.length }</p>
      <p>Current theme: ${props.settings.theme }</p>
      <p>Currency: ${props.settings.currency }</p>
    </div>
  </div>

  <!-- 2. Conditionals Section -->
  <div class="section">
    <h2 class="section-title">2. Conditionals</h2>

    <!-- Simple if condition -->
    <div class="card">
      <h3>User Authentication Status</h3><template x-if="isLoggedIn">
        <p>You are currently logged in as ${props.user.name }.</p>
        <p>Email: ${props.user.email }</p></template><template x-else>
        <p>You are not logged in. Please sign in to access your account.</p></template>
    </div>

    <!-- if/else if/else condition -->
    <div class="card">
      <h3>User Role Check</h3><template x-if="user.role === "admin"">
        <p>Welcome, Administrator! You have full access to the system.</p></template><template x-else-if="user.role === "manager"">
        <p>Welcome, Manager! You can manage products and users.</p></template><template x-else-if="user.role === "editor"">
        <p>Welcome, Editor! You can edit product information.</p></template><template x-else>
        <p>Welcome, User! You have standard user privileges.</p></template>
    </div>

    <!-- Nested conditionals -->
    <div class="card">
      <h3>Product Availability</h3><template x-if="filteredProducts.length > 0">
        <p>We have ${props.filteredProducts.length } products matching your filters:</p><template x-if="settings.filters.inStockOnly">
          <p>Showing only in-stock items.</p></template><template x-else>
          <p>Showing all items (including out of stock).</p></template></template><template x-else>
        <p>No products match your current filters.</p></template>
    </div>
  </div>

  <!-- 3. Loops Section -->
  <div class="section">
    <h2 class="section-title">3. Loops</h2>

    <!-- Simple array loop -->
    <div class="card">
      <h3>Product List</h3>
      <ul><template x-for="product,  in filteredProducts">
          <li>${props.product.name} - ${props.formatPrice(product.price) }<template x-if="!product.inStock">
              <span style="color: red;"> (Out of Stock)</span></template>
          </li></template>
      </ul>
    </div>

    <!-- Loop with index -->
    <div class="card">
      <h3>Numbered Product List</h3>
      <ol><template x-for="product, index in filteredProducts">
          <li value="${index + 1 }">
            ${props.product.name} - ${props.formatPrice(product.price)}<template x-if="product.featured">
              <span style="color: gold;"> ★ Featured</span></template>
          </li></template>
      </ol>
    </div>

    <!-- Object loop -->
    <div class="card">
      <h3>Settings Configuration</h3>
      <dl><template x-for="key, value in settings">
          <dt>${props.key }:</dt><template x-if="typeof value === 'object'">
            <dd>
              <dl><template x-for="subKey, subValue in value">
                  <dt>${props.subKey}:</dt>
                  <dd>${props.subValue}</dd></template>
              </dl>
            </dd></template><template x-else>
            <dd>${props.value}</dd></template></template>
      </dl>
    </div>

    <!-- Nested loops -->
    <div class="card">
      <h3>Categories and Products</h3><template x-for="category,  in categories">
        <div style="margin-bottom: 1rem;">
          <h4>${props.category.name} (${props.category.items.length} items)</h4><template x-if="category.items.length > 0">
            <ul><template x-for="item,  in category.items">
                <li>
                  ${item.name} - ${props.formatPrice(item.price)}
                  <div>
                    Tags:
                    <template x-for="tag,  in item.tags">
                      <span class="tag ${props.getTagClass(tag) }">${props.tag}</span></template>
                  </div>
                </li></template>
            </ul></template><template x-else>
            <p>No items in this category.</p></template>
        </div></template>
    </div>
  </div>

  <!-- 4. Components Section -->
  <div class="section">
    <h2 class="section-title">4. Components</h2>

    <!-- Static components with props -->
    <div class="card">
      <h3>User Profile Component</h3>
    </div>

    <!-- Component loop -->
    <div class="card">
      <h3>Product Cards</h3>
      <div class="product-grid"><template x-for="product,  in filteredProducts.filter(p => p.featured)"></template>
      </div>
    </div>

    <!-- Conditional components -->
    <div class="card">
      <h3>Notifications</h3><template x-for="notification,  in notifications"></template>
    </div>

    <!-- Dynamic components -->
    <div class="card">
      <h3>Dynamic Component Example</h3><template x-if="user.role === "admin""></template><template x-else></template>
    </div>
  </div>

  <!-- 5. Advanced Features Section -->
  <div class="section">
    <h2 class="section-title">5. Advanced Features</h2>

    <!-- Computed values and functions -->
    <div class="card">
      <h3>Filtered Products (Computed)</h3>
      <p>Products between ${props.settings.filters.minPrice } and ${props.settings.filters.maxPrice}:</p>
      <ul><template x-for="product,  in filteredProducts">
          <li>${props.product.name} - ${props.formatPrice(product.price)}</li></template>
      </ul>
    </div>

    <!-- Complex expressions -->
    <div class="card">
      <h3>Complex Expressions</h3>
      <p>Average product price: ${props.formatPrice(products.reduce((props.sum, p) => props.sum + p.price, 0) / props.products.length)}</p>
      <p>Featured products: ${props.products.filter(p => p.featured).length}</p>
      <p>In-stock products: ${props.products.filter(p => p.inStock).length}</p>
      <p>Total inventory value: ${props.formatPrice(products.filter(p => p.inStock).reduce((sum, p) => sum + p.price, 0))}</p>
    </div>

    <!-- Combining features -->
    <div class="card">
      <h3>Combined Features Example</h3><template x-if="filteredProducts.length > 0">
        <div>
          <p>Top ${Math.min(3, filteredProducts.length)} products:</p>
          <div class="product-grid"><template x-for="product, index in filteredProducts.slice(0, 3)">
              <div class="card">
                <h4>${index + 1}. ${props.product.name}</h4>
                <p>Price: ${props.formatPrice(product.price)}</p>
                <div><template x-if="product.inStock">
                    <span style="color: green;">In Stock</span></template><template x-else>
                    <span style="color: red;">Out of Stock</span></template>
                </div>
                <div>
                  Tags:
                  <template x-for="tag,  in product.tags">
                    <span class="tag ${props.getTagClass(tag) }">${props.tag}</span></template>
                </div>
              </div></template>
          </div>
        </div></template><template x-else>
        <p>No products available.</p></template>
    </div>
  </div>

  <!-- Footer component -->
</div>

</body>
</html>`,
  '../content/jim-test.html': (props) => `<html lang="en">
<body>
  <!-- CMS Panel -->
  <div id="plenti_cms"></div>
  <button id="toggle_plenti_cms">Toggle CMS</button>

  <main style="padding: 2rem;">
     <div style="background-color: #f0f0f0; padding: 0.5rem 1rem; border-radius: 0.25rem; margin-bottom: 1rem; font-size: 0.875rem; color: #666;">
      ⚡ Build time: <strong>${props.buildTime}</strong>
    </div>

    <h1>${props.salutation} ${props.name}!</h1>

    <!-- Basic conditionals --><template x-if="name.length > 3">
      <div id="praise">${props.name} is a long name</div><template x-if="age > 1">
        <div>Has been born</div></template></template><template x-else-if="name.length == 2">
      <div id="praise">${props.name} is medium</div></template><template x-else>
      <div id="praise">${props.name} is a short name</div></template>

    <!-- Component usage with props --><template x-if="age > 0">
      <div style="margin: 2rem 0;">
        <h2>Age Component Examples</h2>
        <div style="margin-bottom: 0.5rem; font-size: 0.875rem; color: #666;">Passing dynamic props to components:</div>
      </div>

      <!-- Dynamic components (Jim's innovative <= syntax) with object props -->
      <div style="margin: 2rem 0;">
        <h2>User Profile Examples with Object Props</h2>
      </div>

      <!-- Loops with nested conditionals and advanced JS expressions -->
      <div class="animals" style="background-color: #f0f0f0; padding: 1rem; margin: 2rem 0;">
        <h2>Animals Loop with Advanced Features</h2><template x-for="animal,  in animals"><template x-if="animal == "cat"">
            <div>Hi ${props.animal}!</div></template><template x-else>
            <div>Bye ${props.animal}.</div></template>
          <div class="type-${props.animal}">${props.name} likes: ${props.animal}s</div>
          <div style="color: #666; font-size: 0.875rem;">Backwards: ${props.animal.split('').reverse().join('')}</div>
          <button onclick="${props.animals = props.animals.filter(a => a !== animal)}">Remove ${props.animal}</button>
          <br></br></template>
        <h3>Add new animal:</h3>
        <input type="text" name="newAnimal" x-model="newAnimal" placeholder="animal name"></input>
        <button onclick="${props.animals = [props.newAnimal, props....animals]props.; props.newAnimal = ''}">Add Animal</button>
      </div>

      <!-- Advanced loop examples: array spread, different variable declarations -->
      <div style="background-color: #f9fafb; border: 1px solid #e5e7eb; border-radius: 0.5rem; padding: 1.5rem; margin: 2rem 0;">
        <h2>Advanced Loop Patterns</h2>

        <h3 style="font-size: 1rem; margin: 1rem 0 0.5rem; color: #374151;">Array Spread in Loop:</h3>
        <div style="background-color: white; padding: 1rem; border-radius: 0.375rem; margin-bottom: 1.5rem;"><template x-for="animal,  in ["🦄 unicorn", ...animals]">
            <div style="padding: 0.25rem 0; color: #1f2937;">${props.animal}</div></template>
        </div>

        <h3 style="font-size: 1rem; margin: 1rem 0 0.5rem; color: #374151;">Inline Array Iteration:</h3>
        <div style="background-color: white; padding: 1rem; border-radius: 0.375rem;"><template x-for="word,  in ["Waller", "loves Plenti", "uses AI", "is Australian"]">
            <div style="padding: 0.25rem 0; color: #1f2937;">${props.name} ${props.word}</div></template><template x-for="phrase,  in ["rocks", "codes", "innovates"]">
            <div style="padding: 0.25rem 0; color: #1f2937; font-style: italic;">${props.name} ${props.phrase}!</div></template>
        </div>
      </div>

      <!-- Dynamic Todos Component Examples -->
      <div style="margin: 2rem 0;">
        <h2>Task List - First 5 Tasks</h2>

        <h2 style="margin-top: 2rem;">Task List - Tasks 6-14</h2>
      </div>

      <!-- Admin Panel Component Test -->
      <div style="margin: 2rem 0;">
        <h2>Admin Dashboard Panel</h2>
      </div>

      <!-- Interactive Notification Examples -->
      <div style="margin: 2rem 0;">
        <h2>Interactive Notification Examples</h2>
        <div style="display: flex; gap: 0.5rem; margin-bottom: 1rem; flex-wrap: wrap;"><template x-for="notif,  in notifications">
            <button style="padding: 0.5rem 1rem; border-radius: 0.375rem; border: none; cursor: pointer; background-color: #3b82f6; color: white;" onclick="${props.currentNotification = props.notif}">
              Show ${props.notif.type}
            </button></template>
          <button style="padding: 0.5rem 1rem; border-radius: 0.375rem; border: none; cursor: pointer; background-color: #6b7280; color: white;" onclick="${props.currentNotification = props.null}">
            Clear
          </button>
        </div><template x-if="currentNotification"></template>
      </div></template>
  </main>
</body>
</html>`,
  '../components/services2437.html': (props) => `<!-- ============================================ --><!--                  Services                    --><!-- ============================================ --><section id="services-2437">
	<div class="cs-container">
		<div class="cs-content">
			<span class="cs-topper">Main Services</span>
			<h2 class="cs-title">Explore More Ways to Glow and Express Yourself</h2>
		</div>
		<ul class="cs-card-group">
			<li class="cs-item">
				<a href="" class="cs-link">
					<h3 class="cs-h3">Custom UV Body Painting</h3>
					<picture class="cs-picture">
						<!--Mobile Image-->
						<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art5.jpg">
						<!--Tablet and above Image-->
						<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art5.jpg">
						<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art5.jpg" alt="tattoo" width="373" height="420">
					</picture>
					<p class="cs-item-text">
                        High-definition, blacklight-reactive designs for events, portraits, or personal expression.
                    </p>
                    <div class="cs-border cs-border1" aria-hidden="true"></div>
                    <div class="cs-border cs-border2" aria-hidden="true"></div>
				</a>
			</li>
			<li class="cs-item">
				<a href="" class="cs-link">
					<h3 class="cs-h3">Spiritual & Symbolic Art</h3>
					<picture class="cs-picture">
						<!--Mobile Image-->
						<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art6.jpg">
						<!--Tablet and above Image-->
						<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art6.jpg">
						<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art6.jpg" alt="room" width="373" height="420">
					</picture>
					<p class="cs-item-text">
                        Sacred geometry, mythic creatures, abstract motifs, and personal symbols
                    </p>
                    <div class="cs-border cs-border1" aria-hidden="true"></div>
                    <div class="cs-border cs-border2" aria-hidden="true"></div>
				</a>
			</li>
			<li class="cs-item">
				<a href="" class="cs-link">
					<h3 class="cs-h3">Workshops & Education</h3>
					<picture class="cs-picture">
						<!--Mobile Image-->
						<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art7.jpg">
						<!--Tablet and above Image-->
						<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art7.jpg">
						<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art7.jpg" alt="tattoo" width="373" height="420">
					</picture>
					<p class="cs-item-text">
                        Beginner-friendly body art training and creative process sharing
                    </p>
                    <div class="cs-border cs-border1" aria-hidden="true"></div>
                    <div class="cs-border cs-border2" aria-hidden="true"></div>
				</a>
			</li>
            <li class="cs-item">
				<a href="" class="cs-link">
					<h3 class="cs-h3">Workshops & Education</h3>
					<picture class="cs-picture">
						<!--Mobile Image-->
						<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art8.jpg">
						<!--Tablet and above Image-->
						<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art8.jpg">
						<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art8.jpg" alt="tattoo" width="373" height="420">
					</picture>
					<p class="cs-item-text">
                        Beginner-friendly body art training and creative process sharing
                    </p>
                    <div class="cs-border cs-border1" aria-hidden="true"></div>
                    <div class="cs-border cs-border2" aria-hidden="true"></div>
				</a>
			</li>
		</ul>
        <a href="" class="cs-button-solid">Explore All Services</a>
	</div>
    <picture class="cs-background">
		<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/Graphics/paint.png">
		<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/Graphics/paint.png">
		<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/Graphics/paint.png" alt="splatter" width="1920" height="1168" aria-hidden="true">
	</picture>
</section>`,
  'Nav-test': (props) => `<html lang="en">
<head>
	<meta charset="UTF-8"></meta>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
	<title>${props.title}</title>

	<!-- Alpine.js - Global reactive framework -->
	<script src="//unpkg.com/alpinejs" defer=""></script>

	<!-- App Styles -->
	<link rel="stylesheet" href="/styles/style.css"></link>
</head>
<body>

<!-- Import and render Nav component -->

<!-- Test content -->
<div style="padding: 2rem; max-width: 800px; margin: 0 auto;">
	<h1>Navigation Component Test</h1>
	<p>If you see the navigation above, the Nav component is working correctly!</p>

	<h2>Test Checklist:</h2>
	<ul>
		<li>✓ Navigation bar appears at top</li>
		<li>✓ Logo is visible</li>
		<li>✓ Menu items (Home, About, Services, Artists, FAQ, Contact) are visible</li>
		<li>✓ Dark mode toggle button appears</li>
		<li>✓ Dropdown menus work (Services, Artists)</li>
		<li>✓ Mobile hamburger menu works on small screens</li>
	</ul>

	<h2>How to Test:</h2>
	<ol>
		<li>Resize your browser window to test responsive behavior</li>
		<li>Click on "Services" or "Artists" to test dropdown menus</li>
		<li>Click the dark mode toggle (moon/sun icon)</li>
		<li>On mobile width, click the hamburger menu (three lines)</li>
	</ol>
</div>

</body>
</html>`,
  'Store-demo': (props) => `<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Global Store System - Component Demo</title>
    <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
  </head>
  <body :style="\`background-color: $${$store.theme.getCurrentColors().background}; color: $${$store.theme.getCurrentColors().text}; --code-bg: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#5a7ec5'}; transition: all 0.3s ease;\`">
    <div class="container">
      <h1 :style="\`color: $${$store.theme.getCurrentColors().text};\`">
        Global Store System - Component Demo
      </h1>
      <p class="subtitle" :style="\`color: $${$store.theme.getCurrentColors().secondary};\`">
        Reusable components demonstrating the Global Store System
      </p>

      <!-- Header with Store Components -->
      <div class="header" :style="\`background: ${$store.theme.mode === 'light' ? 'white' : '#2d3748'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
      </div>

      <!-- Interactive Demo Section -->
      <div class="demo-section" :style="\`background: ${$store.theme.mode === 'light' ? 'white' : '#2d3748'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
        <h2>🧪 Try the Interactive Demo</h2>
        <p>
          These buttons modify the stores, and all components above update
          automatically:
        </p>

        <!-- Auth Actions -->
        <div class="action-group" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>Authentication Actions</h3>
          <button @click="$store.auth.login()" class="btn btn-primary">
            Login as Test User
          </button>
          <button @click="$store.auth.logout()" class="btn btn-secondary">
            Logout
          </button>
        </div>

        <!-- Cart Actions -->
        <div class="action-group" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>Shopping Cart Actions</h3>
          <button @click="$store.cart.addItem({ name: 'Widget', price: 9.99 })" class="btn btn-success">
            Add Widget ($9.99)
          </button>
          <button @click="$store.cart.addItem({ name: 'Gadget', price: 19.99 })" class="btn btn-success">
            Add Gadget ($19.99)
          </button>
          <button @click="$store.cart.addItem({ name: 'Doohickey', price: 29.99 })" class="btn btn-success">
            Add Doohickey ($29.99)
          </button>
          <button @click="$store.cart.clear()" class="btn btn-danger">
            Clear Cart
          </button>
        </div>

        <!-- Theme Actions -->
        <div class="action-group" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>Theme Actions</h3>
          <p class="note">
            The ThemeToggle component above provides its own buttons, but you
            can also control the theme from anywhere in your app:
          </p>
          <button @click="$store.theme.setLight()" class="btn btn-light">
            Set Light Mode
          </button>
          <button @click="$store.theme.setDark()" class="btn btn-dark">
            Set Dark Mode
          </button>
          <div style="
              margin-top: 1rem;
              padding: 1rem;
              background: rgba(59, 130, 246, 0.1);
              border-radius: 6px;
              border-left: 4px solid #3b82f6;
            ">
            <strong>Visual Feedback Test:</strong>
            <p style="margin-top: 0.5rem; font-size: 0.95rem">
              Click the theme buttons above and watch the entire page change
              colors! Current mode: <strong x-text="$store.theme.mode"></strong>
            </p>
          </div>
        </div>
      </div>

      <!-- Store State Display -->
      <div class="state-section" :style="\`background: ${$store.theme.mode === 'light' ? 'white' : '#2d3748'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
        <h2>📊 Current Store State</h2>
        <p>Real-time view of the store data:</p>

        <div class="state-grid">
          <!-- Auth Store State -->
          <div class="state-card store-card-hover" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
            <h3>🔐 Auth Store</h3><template x-if="$auth.isLoggedIn">
            <p><strong>Status:</strong> ✅ Logged In</p>
            <p><strong>User:</strong> </p>
            <p><strong>Email:</strong> </p></template><template x-else>
            <p><strong>Status:</strong> ❌ Logged Out</p>
            <p><em>Please log in to access your account</em></p></template>
            <small style="color: #6b7280">💡 Hover to see store object</small>
            <div class="tech-details">
              <pre class="state-display">isLoggedIn: <span x-text="$store.auth.isLoggedIn" style="color: #10b981;"></span>
user: <span x-text="$store.auth.user ? $store.auth.user.name : 'null'" style="color: #10b981;"></span></pre>
            </div>
          </div>

          <!-- Cart Store State -->
          <div class="state-card store-card-hover" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
            <h3>🛒 Cart Store</h3><template x-if="$cart.items.length > 0">
            <p><strong>Items:</strong> </p>
            <p><strong>Total:</strong> $</p>
            <div class="cart-items-list">
              <strong>Cart Contents:</strong>
              <ul><template x-for="item,  in $cart.items">
                <li>${item.name} - $${item.price}</li></template>
              </ul>
            </div></template><template x-else>
            <p><strong>Items:</strong> 0</p>
            <p><strong>Total:</strong> $0.00</p>
            <p><em>Cart is empty</em></p></template>
            <small style="color: #6b7280">💡 Hover to see store object</small>
            <div class="tech-details">
              <pre class="state-display">items: <span x-text="'[' + $store.cart.items.length + ' items]'" style="color: #10b981;"></span>
total: $<span x-text="$store.cart.formattedTotal" style="color: #10b981;"></span></pre>
            </div>
          </div>

          <!-- Theme Store State -->
          <div class="state-card store-card-hover" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
            <h3>🎨 Theme Store</h3><template x-if="$theme.mode === 'light'">
            <p><strong>Mode:</strong> ☀️ Light</p></template><template x-else>
            <p><strong>Mode:</strong> 🌙 Dark</p></template>
            <p>
              <strong>Background:</strong>
              <span x-text="$store.theme.getCurrentColors().background"></span>
            </p>
            <p>
              <strong>Text Color:</strong>
              <span x-text="$store.theme.getCurrentColors().text"></span>
            </p>
            <small style="color: #6b7280">💡 Hover to see store object</small>
            <div class="tech-details">
              <pre class="state-display">mode: "<span x-text="$store.theme.mode" style="color: #10b981;"></span>"
background: "<span x-text="$store.theme.getCurrentColors().background" style="color: #10b981;"></span>"
text: "<span x-text="$store.theme.getCurrentColors().text" style="color: #10b981;"></span>"</pre>
            </div>
          </div>
        </div>
      </div>

      <!-- Documentation Section -->
      <div class="docs-section" :style="\`background: ${$store.theme.mode === 'light' ? 'white' : '#2d3748'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
        <h2>📚 How It Works</h2>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>1. Store Files</h3>
          <p>Stores are defined in <code>stores/</code> directory:</p>
          <ul>
            <li><code>stores/auth.js</code> - Authentication state</li>
            <li><code>stores/cart.js</code> - Shopping cart state</li>
            <li><code>stores/theme.js</code> - Theme preferences</li>
          </ul>
        </div>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>2. Components Reference Stores</h3>
          <p>
            Components declare which stores they use in their fence section:
          </p>
          <pre class="code-display">
---
import store from './stores/auth.js'
---

<template x-if="$auth.isLoggedIn">
  Welcome, !</template></pre>
        </div>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>3. Automatic Reactivity</h3>
          <p>When any component modifies a store:</p>
          <ol>
            <li>Alpine.js detects the change</li>
            <li>All components using that store automatically update</li>
            <li>No manual event handling needed!</li>
          </ol>
        </div>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>4. Global Scope</h3>
          <p>Access stores anywhere with <code>$store.storeName</code>:</p>
          <ul>
            <li><code>@click="$store.auth.login()"</code> - Call methods</li>
            <li><code></code> - Read properties</li>
            <li>
              <code>{if $cart.items.length > 0}</code> - Use in conditionals
            </li>
            <li><code>{for item in $cart.items}</code> - Iterate arrays</li>
          </ul>
        </div>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'}; border-left: 4px solid #10b981;\`">
          <h3 style="color: #10b981">5. Theme Reactivity</h3>
          <p>
            Notice how all elements on this page use
            <code>:style</code> bindings:
          </p>
          <pre class="code-display">
:style="&#96;background: &#36;&#123;&#36;store.theme.getCurrentColors().background&#125;;
         color: &#36;&#123;&#36;store.theme.getCurrentColors().text&#125;;&#96;"</pre>
          <p style="margin-top: 0.5rem">
            This makes the entire page reactive to theme changes!
          </p>
        </div>
      </div>
    </div>
  </body>
</html>`,
  '../content/store-test-with-theme.html': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>${props.title}</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body :style="$store.theme.bodyStyle">

  <h1 style="text-align: center;">${props.title}</h1>
  <p style="text-align: center;">Current theme: <strong></strong></p>

  <!-- APPROACH 1: Inline Styles in Store (using getters) -->
  <div :style="$store.theme.sectionStyle">
    <h2 style="margin-top: 0;">🎨 Approach 1: Inline Styles in Store (Getters)</h2>
    <p><strong>Pattern:</strong> Store defines <code>get bodyStyle()</code>, <code>get sectionStyle()</code>, <code>get buttonStyle()</code></p>
    <p><strong>Usage:</strong> <code>:style="$store.theme.buttonStyle"</code></p>
    <button @click="$store.theme.toggle()" :style="$store.theme.buttonStyle">
      Toggle to  Mode
    </button>
    <p style="margin-top: 1rem;">✅ <strong>Benefits:</strong> Single source of truth, DRY, all theme logic in the store</p>
    <p>✅ <strong>Clean HTML:</strong> No duplication of dark/light values in templates</p>
  </div>

  <!-- APPROACH 2: CSS Classes -->
  <div class="theme-section" :class="$store.theme.mode">
    <h2 style="margin-top: 0;">🎨 Approach 2: CSS Classes with :class Binding</h2>
    <p><strong>Pattern:</strong> CSS defines <code>.theme-section.light</code> and <code>.theme-section.dark</code></p>
    <p><strong>Usage:</strong> <code>:class="$store.theme.mode"</code></p>
    <button class="theme-btn" :class="$store.theme.mode" @click="$store.theme.toggle()">
      Toggle to  Mode
    </button>
    <p style="margin-top: 1rem;">✅ <strong>Benefits:</strong> Separation of concerns, CSS in stylesheets, better performance</p>
    <p>✅ <strong>Production Ready:</strong> Standard CSS approach, easier for designers</p>
  </div>

  <!-- Auth Section -->
  <div :style="$store.theme.mode === 'dark' ? 'background: #2d2d2d; padding: 1.5rem; margin-bottom: 1rem; border-radius: 8px;' : 'background: #f5f5f5; padding: 1.5rem; margin-bottom: 1rem; border-radius: 8px;'">
    <h2 style="margin-top: 0;">✅ Authentication Store</h2>
    <p>Status: <strong></strong></p>
    <p>User: <strong></strong></p>
    <p>Email: <strong></strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.auth.login()" :style="$store.theme.mode === 'dark' ? 'background: #4a9eff; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem;' : 'background: #2563eb; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem;'">
        Login
      </button>
      <button @click="$store.auth.logout()" :style="$store.theme.mode === 'dark' ? 'background: #666; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer;' : 'background: #6b7280; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer;'">
        Logout
      </button>
    </div>
  </div>

  <!-- Cart Section -->
  <div :style="$store.theme.mode === 'dark' ? 'background: #2d2d2d; padding: 1.5rem; margin-bottom: 1rem; border-radius: 8px;' : 'background: #f5f5f5; padding: 1.5rem; margin-bottom: 1rem; border-radius: 8px;'">
    <h2 style="margin-top: 0;">🛒 Shopping Cart Store</h2>
    <p>Items in cart: <strong></strong></p>
    <p>Total: <strong>$</strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.cart.addItem('Widget', 9.99)" :style="$store.theme.mode === 'dark' ? 'background: #4a9eff; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;' : 'background: #2563eb; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;'">
        Add Widget ($9.99)
      </button>
      <button @click="$store.cart.addItem('Gadget', 19.99)" :style="$store.theme.mode === 'dark' ? 'background: #4a9eff; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;' : 'background: #2563eb; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;'">
        Add Gadget ($19.99)
      </button>
      <button @click="$store.cart.addItem('Doohickey', 29.99)" :style="$store.theme.mode === 'dark' ? 'background: #4a9eff; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;' : 'background: #2563eb; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;'">
        Add Doohickey ($29.99)
      </button>
      <button @click="$store.cart.clear()" :style="$store.theme.mode === 'dark' ? 'background: #dc2626; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer;' : 'background: #dc2626; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer;'">
        Clear Cart
      </button>
    </div>
  </div>

  <!-- Comparison Section -->
  <div :style="$store.theme.sectionStyle">
    <h2 style="margin-top: 0;">📊 Approach Comparison</h2>
    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th style="text-align: left; padding: 0.5rem; border-bottom: 2px solid #ccc;">Aspect</th>
          <th style="text-align: left; padding: 0.5rem; border-bottom: 2px solid #ccc;">Approach 1: Store Getters</th>
          <th style="text-align: left; padding: 0.5rem; border-bottom: 2px solid #ccc;">Approach 2: CSS Classes</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>Code Location</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">Styles in store (fence)</td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">Styles in <code>&lt;style&gt;</code> tag</td>
        </tr>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>HTML Usage</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><code>:style="$store.theme.buttonStyle"</code></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><code>:class="$store.theme.mode"</code></td>
        </tr>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>DRY Principle</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">✅ Excellent - single source</td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">✅ Excellent - CSS only</td>
        </tr>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>Performance</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">Good - inline styles</td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">✅ Better - CSS caching</td>
        </tr>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>Designer Friendly</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">JavaScript knowledge needed</td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">✅ Standard CSS workflow</td>
        </tr>
        <tr>
          <td style="padding: 0.5rem;"><strong>Best For</strong></td>
          <td style="padding: 0.5rem;">Dynamic, complex theming logic</td>
          <td style="padding: 0.5rem;">✅ Production apps, team workflows</td>
        </tr>
      </tbody>
    </table>
  </div>

</body>
</html>`,
  'Store-test': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>${props.title}</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <style>
    body 
      font-family: system-ui, -apple-system, sans-serif;
      max-width: 800px;
      margin: 0 auto;
      padding: 2rem;
      line-height: 1.6;
    
    .section 
      background: #f5f5f5;
      padding: 1.5rem;
      margin-bottom: 1.5rem;
      border-radius: 8px;
    
    .section h2 
      margin-top: 0;
      color: #333;
    
    button 
      background: #0066cc;
      color: white;
      border: none;
      padding: 0.5rem 1rem;
      border-radius: 4px;
      cursor: pointer;
      margin-right: 0.5rem;
    
    button:hover 
      background: #0052a3;
    
    .user-info 
      background: white;
      padding: 1rem;
      border-radius: 4px;
      margin-top: 1rem;
    
    .cart-item 
      background: white;
      padding: 0.5rem;
      margin-bottom: 0.5rem;
      border-radius: 4px;
    
    .status 
      font-weight: bold;
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
    
    .logged-in  background: #d4edda; color: #155724; 
    .logged-out  background: #f8d7da; color: #721c24; 
  </style>
</head>
<body>
  <h1>${props.title}</h1>
  <p>Testing Task 2.1: Store Expression Transformer</p>

  <!-- Test 1: Store expressions in text content -->
  <div class="section">
    <h2>Test 1: Text Content with Store Expressions</h2>
    <p>Auth Status: <span class="status"></span></p>
    <p>User Name: </p>
    <p>User Email: </p>
    <p>Cart Items: </p>
    <p>Cart Total: $</p>
    <p>Current Theme: </p>
  </div>

  <!-- Test 2: Store expressions in attributes (Alpine directives) -->
  <div class="section">
    <h2>Test 2: Store Expressions in Attributes</h2>
    <button @click="$store.auth.login()">Login</button>
    <button @click="$store.auth.logout()">Logout</button>

    <div class="user-info">
      <p><strong>Logged In:</strong> <span></span></p>
      <p><strong>Name:</strong> </p>
      <p><strong>Email:</strong> </p>
    </div>
  </div>

  <!-- Test 3: Cart functionality -->
  <div class="section">
    <h2>Test 3: Cart Store</h2>
    <button @click="$store.cart.addItem('Widget', 9.99)">Add Widget ($9.99)</button>
    <button @click="$store.cart.addItem('Gadget', 19.99)">Add Gadget ($19.99)</button>
    <button @click="$store.cart.addItem('Doohickey', 29.99)">Add Doohickey ($29.99)</button>
    <button @click="$store.cart.clear()">Clear Cart</button>

    <div style="margin-top: 1rem;">
      <p><strong>Items in cart:</strong> </p>
      <p><strong>Total:</strong> $</p>
    </div>
  </div>

  <!-- Test 4: Theme toggle -->
  <div class="section">
    <h2>Test 4: Theme Store</h2>
    <button @click="$store.theme.toggle()">Toggle Theme</button>
    <p>Current theme mode: <strong></strong></p>
  </div>

  <!-- Test 5: Multiple stores in one expression -->
  <div class="section">
    <h2>Test 5: Combined Store Data</h2>
    <p>User  has  items in cart using  theme</p>
  </div>

  <!-- Debug Info -->
  <div class="section" style="background: #fff3cd;">
    <h2>Debug Info</h2>
    <p><strong>What we're testing:</strong></p>
    <ul>
      <li>Store definitions in fence section parse correctly</li>
      <li>Store expressions like  transform to Alpine.js syntax</li>
      <li>Nested properties like  work</li>
      <li>Store methods can be called from Alpine directives</li>
    </ul>
    <p><strong>Expected behavior:</strong></p>
    <ul>
      <li>Initial state shows "false", "Guest", empty cart</li>
      <li>Login button updates auth store and displays user info</li>
      <li>Add item buttons update cart count and total</li>
      <li>Theme toggle switches between "light" and "dark"</li>
    </ul>
  </div>
</body>
</html>`,
  '../components/CartBadge.html': (props) => `<div class="cart-badge">
  <div class="cart-icon-wrapper">
    🛒
    <template x-if="$cart.items.length > 0">
      <span class="cart-count"></span></template>
  </div>
  <div class="cart-details">
    <div class="cart-items"><template x-if="$cart.items.length === 0">
        <span class="empty-cart">Cart is empty</span></template><template x-else>
        <span class="item-count"> items</span></template>
    </div><template x-if="$cart.items.length > 0">
      <div class="cart-total">
        Total: $<strong x-text="$store.cart.formattedTotal"></strong>
      </div></template>
  </div>
</div>`,
  '../components/ThemeToggle.html': (props) => `<div class="theme-toggle">
  <span class="theme-label">Theme:</span>
  <div class="toggle-buttons">
    <button @click="$store.theme.setLight()" class="theme-btn" :class="{ active: $store.theme.mode === 'light' }">
      ☀️ Light
    </button>
    <button @click="$store.theme.setDark()" class="theme-btn" :class="{ active: $store.theme.mode === 'dark' }">
      🌙 Dark
    </button>
  </div>
  <button @click="$store.theme.toggle()" class="toggle-btn" title="Toggle theme">
    🔄
  </button>
</div>`,
  'Age': (props) => `<div class="age-container">
  <div class="age-info">${props.name}'s age is: ${props.age}</div><template x-if="age > 300">
    <div class="age-badge dust">Dust (300+)</div></template><template x-else-if="age > 200">
    <div class="age-badge dead">Dead (200+)</div></template><template x-else-if="age > 90">
    <div class="age-badge old">Wow (90+)</div></template><template x-else-if="age > 80">
    <div class="age-badge old">Well done (80+)</div></template><template x-else-if="age > 40">
    <div class="age-badge mid">Over the hill (40+)</div></template><template x-else-if="age >= 21">
    <div class="age-badge young">Can drink (21+)</div></template><template x-else>
    <div class="age-badge new">Just a baby (Under 21)</div></template>
</div>`,
  '../components/hero2436.html': (props) => `<!-- ============================================ --><!--                   Hero                       --><!-- ============================================ --><section id="hero-2436">
    <div class="cs-container">
        <div class="cs-content">
            <div class="cs-flex cs-flex1">
                <span class="cs-topper">${props.topper}</span>
                <h2 class="cs-title">${props.title}</h2>
            </div>
            <div class="cs-flex cs-flex2">
                <p class="cs-text">
                    ${props.description}
                </p>
                <a href="${props.buttonLink}" class="cs-button-solid">${props.buttonText}</a>
            </div>
        </div>
        <img class="cs-logo-feature" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/Logos/art-logo.svg" alt="logo" width="1280" height="239"></img>
    </div>
    <!--Background Image-->
    <picture class="cs-background">
        <!--Mobile Image-->
        <source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art4.jpg"></source>
        <!--Tablet and above Image-->
        <source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art4.jpg"></source>
        <img decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art4.jpg" alt="face with body art" width="1800" height="860" aria-hidden="true" fetchpriority="high"></img>
    </picture>
</section>`,
  '../components/userdashboard.html': (props) => `<div class="dashboard">
  <div class="dashboard-header">
    <div class="welcome-message">Welcome back, ${props.user.name}!</div>
    <div>${props.user.email}</div>
  </div>
  
  <div class="dashboard-content">
    <div class="dashboard-sections">
      <div>
        <div class="dashboard-section">
          <div class="section-title">
            <span>Your Recent Orders</span>
            <a href="#" class="section-link">View All Orders</a>
          </div><template x-if="user.orders && user.orders.length > 0">
            <table class="order-table">
              <thead>
                <tr>
                  <th>Order ID</th>
                  <th>Date</th>
                  <th>Status</th>
                  <th>Total</th>
                </tr>
              </thead>
              <tbody><template x-for="order,  in user.orders">
                  <tr>
                    <td>${props.order.id }</td>
                    <td>${props.formatDate(order.date)}</td>
                    <td>
                      <span class="status-badge ${props.getStatusColor(order.status) }">
                        ${props.order.status }
                      </span>
                    </td>
                    <td>${props.formatCurrency(order.total)}</td>
                  </tr></template>
              </tbody>
            </table>
            
            <div class="summary-card">
              <div class="summary-row">
                <span class="summary-label">Total Orders</span>
                <span>${props.user.orders.length}</span>
              </div>
              <div class="summary-total">
                <span>Total Spent</span>
                <span>${props.formatCurrency(totalOrdersValue)}</span>
              </div>
            </div></template><template x-else>
            <p>You haven't placed any orders yet.</p></template>
        </div>
        
        <div class="dashboard-actions">
          <button class="dashboard-button">View All Orders</button>
          <button class="dashboard-button dashboard-button-outline">Track an Order</button>
        </div>
      </div>
      
      <div>
        <div class="dashboard-section">
          <div class="section-title">
            <span>Your Wishlist</span>
            <a href="#" class="section-link">View All</a>
          </div><template x-if="user.wishlist && user.wishlist.length > 0">
            <div><template x-for="item,  in user.wishlist">
                <div class="wishlist-item">
                  <div class="wishlist-name">${item.name}</div>
                  <div class="wishlist-price">${props.formatCurrency(item.price)}</div>
                </div></template>
            </div></template><template x-else>
            <p>Your wishlist is empty.</p></template>
        </div>
        
        <div class="dashboard-section">
          <div class="section-title">Account Settings</div>
          <div>
            <button class="dashboard-button dashboard-button-outline" style="width: 100%; margin-bottom: 0.5rem;">
              Edit Profile
            </button>
            <button class="dashboard-button dashboard-button-outline" style="width: 100%; margin-bottom: 0.5rem;">
              Change Password
            </button>
            <button class="dashboard-button dashboard-button-outline" style="width: 100%;">
              Manage Payment Methods
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>`,
  'Userprofile': (props) => `<template x-if="compact">
  <div class="profile-card profile-compact">
    <div class="profile-avatar-small"><template x-if="user.avatar">
        <img src="${props.user.avatar}" alt="${props.user.name}"></template><template x-else-if="user.name">
        ${props.user.name.charAt(0)}</template><template x-else>
        U</template>
    </div>
    <div>
      <div class="profile-name-compact">${props.user.name}</div><template x-if="showRole">
        <span class="profile-role ${props.roleBadge.bg} ${props.roleBadge.text}">
          ${props.user.role}
        </span></template>
    </div>
  </div></template><template x-else>
  <div class="profile-card">
    <div class="profile-header">
      <div class="profile-avatar"><template x-if="user.avatar">
          <img src="${props.user.avatar}" alt="${props.user.name}"></template><template x-else-if="user.name">
          ${props.user.name.charAt(0)}</template><template x-else>
          U</template>
      </div>
      <h2 class="profile-name">${props.user.name}</h2><template x-if="showRole">
        <span class="profile-role ${props.roleBadge.bg} ${props.roleBadge.text}">
          ${props.user.role}
        </span></template>
    </div>
    <div class="profile-details"><template x-if="showEmail">
        <div class="profile-detail">
          <div class="profile-label">Email</div>
          <div class="profile-value">${props.user.email}</div>
        </div></template><template x-if="showJoinDate">
        <div class="profile-detail">
          <div class="profile-label">Member since</div>
          <div class="profile-value">${props.formattedJoinDate}</div>
        </div></template>
    </div>
  </div></template>`,
  '../global/head.html': (props) => `<head>
	<meta charset="UTF-8"></meta>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
	<title>${props.title}</title>
	<meta name="description" content="${props.description}"></meta>

	<!-- Alpine.js - Global reactive framework (MUST load first) -->
	<script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>

	<!-- Runtime Components - Load as module to ensure proper timing -->
	<!-- CRITICAL: Must load BEFORE Alpine initializes, so use immediate execution -->
	<script src="/static/js/runtime-components.js"></script>

	<!-- App Scripts -->
	<script defer="" type="text/javascript" src="/scripts/script.js"></script>
	<script defer="" type="text/javascript" src="/scripts/cms.js"></script>

	<!-- App Styles -->
	<link rel="stylesheet" href="/styles/style.css"></link>
	<link rel="stylesheet" href="/styles/cms.css"></link>
	  <link rel="preload" href="/static/fonts/roboto-v29-latin-regular.woff2 " as="font" type="font/woff2" crossorigin="">
</head>`,
  'Script-test': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <title>Script Tag Preservation Test</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.js"></script>
</head>
<body>
  <h1>Script Tag Preservation Test</h1>

  <div x-data="{ count: ${props.count}, message: '${props.message}' }">
    <p>Count: <span x-text="count"></span></p>
    <p>Message: <span x-text="message"></span></p>
    <button @click="count++">Increment</button>
  </div>

  <script>
    // This inline script should be preserved
    console.log('Script tag preserved successfully!');
    console.log('Initial count:', count);
    console.log('Initial message:', 'message');

    // Client-side JavaScript
    document.addEventListener('DOMContentLoaded', function() 
      console.log('Page loaded with Alpine.js');
    );
  </script>
</body>
</html>`,
  'Userdashboard': (props) => `<div class="dashboard">
  <div class="dashboard-header">
    <div class="welcome-message">Welcome back, ${props.user.name}!</div>
    <div>${props.user.email}</div>
  </div>
  
  <div class="dashboard-content">
    <div class="dashboard-sections">
      <div>
        <div class="dashboard-section">
          <div class="section-title">
            <span>Your Recent Orders</span>
            <a href="#" class="section-link">View All Orders</a>
          </div><template x-if="user.orders && user.orders.length > 0">
            <table class="order-table">
              <thead>
                <tr>
                  <th>Order ID</th>
                  <th>Date</th>
                  <th>Status</th>
                  <th>Total</th>
                </tr>
              </thead>
              <tbody><template x-for="order,  in user.orders">
                  <tr>
                    <td>${props.order.id }</td>
                    <td>${props.formatDate(order.date)}</td>
                    <td>
                      <span class="status-badge ${props.getStatusColor(order.status) }">
                        ${props.order.status }
                      </span>
                    </td>
                    <td>${props.formatCurrency(order.total)}</td>
                  </tr></template>
              </tbody>
            </table>
            
            <div class="summary-card">
              <div class="summary-row">
                <span class="summary-label">Total Orders</span>
                <span>${props.user.orders.length}</span>
              </div>
              <div class="summary-total">
                <span>Total Spent</span>
                <span>${props.formatCurrency(totalOrdersValue)}</span>
              </div>
            </div></template><template x-else>
            <p>You haven't placed any orders yet.</p></template>
        </div>
        
        <div class="dashboard-actions">
          <button class="dashboard-button">View All Orders</button>
          <button class="dashboard-button dashboard-button-outline">Track an Order</button>
        </div>
      </div>
      
      <div>
        <div class="dashboard-section">
          <div class="section-title">
            <span>Your Wishlist</span>
            <a href="#" class="section-link">View All</a>
          </div><template x-if="user.wishlist && user.wishlist.length > 0">
            <div><template x-for="item,  in user.wishlist">
                <div class="wishlist-item">
                  <div class="wishlist-name">${item.name}</div>
                  <div class="wishlist-price">${props.formatCurrency(item.price)}</div>
                </div></template>
            </div></template><template x-else>
            <p>Your wishlist is empty.</p></template>
        </div>
        
        <div class="dashboard-section">
          <div class="section-title">Account Settings</div>
          <div>
            <button class="dashboard-button dashboard-button-outline" style="width: 100%; margin-bottom: 0.5rem;">
              Edit Profile
            </button>
            <button class="dashboard-button dashboard-button-outline" style="width: 100%; margin-bottom: 0.5rem;">
              Change Password
            </button>
            <button class="dashboard-button dashboard-button-outline" style="width: 100%;">
              Manage Payment Methods
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>`,
  'Footer': (props) => `<!-- ============================================ --><!--                   Footer                     --><!-- ============================================ --><footer id="cs-footer-2440">
    <div class="cs-container">
        <div class="cs-top">
            <div class="cs-logo-group">
                <a class="cs-logo" aria-label="go back to home" href="">
                    <img class="cs-logo-img" aria-hidden="true" loading="lazy" decoding="async" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/artistitch.svg" alt="logo" width="260" height="48"></img>
                </a>
                <ul class="cs-contact-group">
                    <li class="cs-contact-li">
                        <img class="cs-contact-icon" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/white-phone.svg" alt="phone icon" height="24" width="24" loading="lazy" decoding="async" aria-hidden="true"></img>
                        <a class="cs-contact-link" href="tel:555-474-7386">555-474-7386</a>
                    </li>
                    <li class="cs-contact-li">
                        <img class="cs-contact-icon" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/white-email.svg" alt="email icon" height="24" width="24" loading="lazy" decoding="async" aria-hidden="true"></img>
                        <a class="cs-contact-link" href="mailto:ArtiStitchart@gmail.com">ArtiStitchart@gmail.com</a>
                    </li>
                    <li class="cs-contact-li">
                        <img class="cs-contact-icon" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/white-location.svg" alt="location icon" height="24" width="24" loading="lazy" decoding="async" aria-hidden="true"></img>
                        <a class="cs-contact-link" href="">555 Fake St, Portland, OR 97217, United States</a>
                    </li>
                </ul>
            </div>
            <!--Sitemap-->
            <ul class="cs-ul">
                <li class="cs-li">
                    <span class="cs-header">Quick Links</span>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Home</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">About</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Services</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Artist</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">FAQ</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Contact</a>
                </li>
            </ul>
            <ul class="cs-ul">
                <li class="cs-li">
                    <span class="cs-header">Our Services</span>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Custom UV Body Painting</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Spiritual & Symbolic Art</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Workshops & Education</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Body Art Reworks</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Live Event Painting</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Symbolic Expression Art</a>
                </li>
            </ul>
            <div class="cs-subscribe-group">
                <span class="cs-header">Sign up to receive the latest news and events from us.</span>
                <form class="cs-form" name="Footer Subscribe" method="post">
                    <input class="cs-input" aria-label="email" name="email" type="email" placeholder="Email Address" required=""></input>
                    <button class="cs-button-solid" name="submit" type="submit" aria-label="submit">
                        <img class="cs-icon" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/white-upper-right-arrow.svg" alt="arrow icon" height="24" width="24" loading="lazy" decoding="async" aria-hidden="true"></img>
                    </button>
                </form>
            </div>
        </div>
        <!--Bottom Copyright And Social-->
        <div class="cs-bottom">
            <div class="cs-flex">
                <a href="" class="cs-link">Facebook</a>
                <div class="cs-divider" aria-hidden="true"></div>
                <a href="" class="cs-link">Instagram</a>
                <div class="cs-divider" aria-hidden="true"></div>
                <a href="" class="cs-link">Tiktok</a>
                <div class="cs-divider" aria-hidden="true"></div>
                <a href="" class="cs-link">Youtube</a>
            </div>
            <span class="cs-copyright-group">
                <span class="cs-copyright">© Copyright 2025 -</span>
                <a class="cs-link" href="">ArtiStitch</a>
            </span>
        </div>
    </div>
</footer>`,
  'Jim-test': (props) => `<html lang="en">
<body>
  <!-- CMS Panel -->
  <div id="plenti_cms"></div>
  <button id="toggle_plenti_cms">Toggle CMS</button>

  <main style="padding: 2rem;">
     <div style="background-color: #f0f0f0; padding: 0.5rem 1rem; border-radius: 0.25rem; margin-bottom: 1rem; font-size: 0.875rem; color: #666;">
      ⚡ Build time: <strong>${props.buildTime}</strong>
    </div>

    <h1>${props.salutation} ${props.name}!</h1>

    <!-- Basic conditionals --><template x-if="name.length > 3">
      <div id="praise">${props.name} is a long name</div><template x-if="age > 1">
        <div>Has been born</div></template></template><template x-else-if="name.length == 2">
      <div id="praise">${props.name} is medium</div></template><template x-else>
      <div id="praise">${props.name} is a short name</div></template>

    <!-- Component usage with props --><template x-if="age > 0">
      <div style="margin: 2rem 0;">
        <h2>Age Component Examples</h2>
        <div style="margin-bottom: 0.5rem; font-size: 0.875rem; color: #666;">Passing dynamic props to components:</div>
      </div>

      <!-- Dynamic components (Jim's innovative <= syntax) with object props -->
      <div style="margin: 2rem 0;">
        <h2>User Profile Examples with Object Props</h2>
      </div>

      <!-- Loops with nested conditionals and advanced JS expressions -->
      <div class="animals" style="background-color: #f0f0f0; padding: 1rem; margin: 2rem 0;">
        <h2>Animals Loop with Advanced Features</h2><template x-for="animal,  in animals"><template x-if="animal == "cat"">
            <div>Hi ${props.animal}!</div></template><template x-else>
            <div>Bye ${props.animal}.</div></template>
          <div class="type-${props.animal}">${props.name} likes: ${props.animal}s</div>
          <div style="color: #666; font-size: 0.875rem;">Backwards: ${props.animal.split('').reverse().join('')}</div>
          <button onclick="${props.animals = props.animals.filter(a => a !== animal)}">Remove ${props.animal}</button>
          <br></br></template>
        <h3>Add new animal:</h3>
        <input type="text" name="newAnimal" x-model="newAnimal" placeholder="animal name"></input>
        <button onclick="${props.animals = [props.newAnimal, props....animals]props.; props.newAnimal = ''}">Add Animal</button>
      </div>

      <!-- Advanced loop examples: array spread, different variable declarations -->
      <div style="background-color: #f9fafb; border: 1px solid #e5e7eb; border-radius: 0.5rem; padding: 1.5rem; margin: 2rem 0;">
        <h2>Advanced Loop Patterns</h2>

        <h3 style="font-size: 1rem; margin: 1rem 0 0.5rem; color: #374151;">Array Spread in Loop:</h3>
        <div style="background-color: white; padding: 1rem; border-radius: 0.375rem; margin-bottom: 1.5rem;"><template x-for="animal,  in ["🦄 unicorn", ...animals]">
            <div style="padding: 0.25rem 0; color: #1f2937;">${props.animal}</div></template>
        </div>

        <h3 style="font-size: 1rem; margin: 1rem 0 0.5rem; color: #374151;">Inline Array Iteration:</h3>
        <div style="background-color: white; padding: 1rem; border-radius: 0.375rem;"><template x-for="word,  in ["Waller", "loves Plenti", "uses AI", "is Australian"]">
            <div style="padding: 0.25rem 0; color: #1f2937;">${props.name} ${props.word}</div></template><template x-for="phrase,  in ["rocks", "codes", "innovates"]">
            <div style="padding: 0.25rem 0; color: #1f2937; font-style: italic;">${props.name} ${props.phrase}!</div></template>
        </div>
      </div>

      <!-- Dynamic Todos Component Examples -->
      <div style="margin: 2rem 0;">
        <h2>Task List - First 5 Tasks</h2>

        <h2 style="margin-top: 2rem;">Task List - Tasks 6-14</h2>
      </div>

      <!-- Admin Panel Component Test -->
      <div style="margin: 2rem 0;">
        <h2>Admin Dashboard Panel</h2>
      </div>

      <!-- Interactive Notification Examples -->
      <div style="margin: 2rem 0;">
        <h2>Interactive Notification Examples</h2>
        <div style="display: flex; gap: 0.5rem; margin-bottom: 1rem; flex-wrap: wrap;"><template x-for="notif,  in notifications">
            <button style="padding: 0.5rem 1rem; border-radius: 0.375rem; border: none; cursor: pointer; background-color: #3b82f6; color: white;" onclick="${props.currentNotification = props.notif}">
              Show ${props.notif.type}
            </button></template>
          <button style="padding: 0.5rem 1rem; border-radius: 0.375rem; border: none; cursor: pointer; background-color: #6b7280; color: white;" onclick="${props.currentNotification = props.null}">
            Clear
          </button>
        </div><template x-if="currentNotification"></template>
      </div></template>
  </main>
</body>
</html>`,
  '../content/script-test.html': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <title>Script Tag Preservation Test</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.js"></script>
</head>
<body>
  <h1>Script Tag Preservation Test</h1>

  <div x-data="{ count: ${props.count}, message: '${props.message}' }">
    <p>Count: <span x-text="count"></span></p>
    <p>Message: <span x-text="message"></span></p>
    <button @click="count++">Increment</button>
  </div>

  <script>
    // This inline script should be preserved
    console.log('Script tag preserved successfully!');
    console.log('Initial count:', count);
    console.log('Initial message:', 'message');

    // Client-side JavaScript
    document.addEventListener('DOMContentLoaded', function() 
      console.log('Page loaded with Alpine.js');
    );
  </script>
</body>
</html>`,
  'Store-test-minimal': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <title>${props.title}</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css"></link>
</head>
<body style="padding: 2rem;">
  <main class="container">
    <h1 style="text-align: center; color: #2563eb;">${props.title}</h1>

    <article>
    <h2>✅ Test 1: Authentication Store</h2>
    <p>Status: <span class="status"></span></p>
    <p>User: <strong></strong></p>
    <p>Email: <strong></strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.auth.login()">Login</button>
      <button @click="$store.auth.logout()">Logout</button>
    </div>

    <div class="info-box">
      <strong>Try it:</strong> Click "Login" and watch the values update reactively!<br></br>
      Store expressions: <code></code>, <code></code>
    </div>
  </div>

  <div class="test-section">
    <h2>🛒 Test 2: Shopping Cart Store</h2>
    <p>Items in cart: <strong></strong></p>
    <p>Total: <strong>$</strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.cart.addItem('Widget', 9.99)">Add Widget ($9.99)</button>
      <button @click="$store.cart.addItem('Gadget', 19.99)">Add Gadget ($19.99)</button>
      <button @click="$store.cart.addItem('Doohickey', 29.99)">Add Doohickey ($29.99)</button>
      <button @click="$store.cart.clear()" class="secondary">Clear Cart</button>
    </div>

    <div class="info-box">
      <strong>Try it:</strong> Add items and watch the count and total update!<br></br>
      Store expressions: <code></code>, <code></code>
    </div>
  </div>

  <div class="test-section">
    <h2>🎨 Test 3: Theme Store</h2>
    <p>Current theme: <strong></strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.theme.toggle()">Toggle Theme</button>
    </div>

    <div class="info-box">
      <strong>Try it:</strong> Click to toggle between "light" and "dark"<br></br>
      Store expression: <code></code>
    </div>
  </div>

  div class="test-section">
    <h2>🎉 Store System Features Working</h2>
    <ul>
      <li>✅ <strong>Inline store definitions</strong> in fence section</li>
      <li>✅ <strong>Store expressions</strong> transform to Alpine.js syntax</li>
      <li>✅ <strong>Store initialization</strong> via <code>Alpine.store()</code></li>
      <li>✅ <strong>Reactive updates</strong> across the page</li>
      <li>✅ <strong>Store methods</strong> callable from Alpine directives</li>
      <li>✅ <strong>Nested properties</strong> like <code>$auth.user.name</code></li>
    </ul>

    <div class="info-box">
      <strong>View Page Source</strong> to see:<br></br>
      • Store initialization script in <code>&lt;head&gt;</code><br></br>
      • Store expressions transformed to <code>&lt;span x-text="$store.X.Y"&gt;</code><br></br>
      • Alpine.js event handlers with <code>@click="$store.X.method()"</code>
      </div>
    </article>
  </main>
</body>
</html>`,
  '../content/store-test.html': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>${props.title}</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <style>
    body 
      font-family: system-ui, -apple-system, sans-serif;
      max-width: 800px;
      margin: 0 auto;
      padding: 2rem;
      line-height: 1.6;
    
    .section 
      background: #f5f5f5;
      padding: 1.5rem;
      margin-bottom: 1.5rem;
      border-radius: 8px;
    
    .section h2 
      margin-top: 0;
      color: #333;
    
    button 
      background: #0066cc;
      color: white;
      border: none;
      padding: 0.5rem 1rem;
      border-radius: 4px;
      cursor: pointer;
      margin-right: 0.5rem;
    
    button:hover 
      background: #0052a3;
    
    .user-info 
      background: white;
      padding: 1rem;
      border-radius: 4px;
      margin-top: 1rem;
    
    .cart-item 
      background: white;
      padding: 0.5rem;
      margin-bottom: 0.5rem;
      border-radius: 4px;
    
    .status 
      font-weight: bold;
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
    
    .logged-in  background: #d4edda; color: #155724; 
    .logged-out  background: #f8d7da; color: #721c24; 
  </style>
</head>
<body>
  <h1>${props.title}</h1>
  <p>Testing Task 2.1: Store Expression Transformer</p>

  <!-- Test 1: Store expressions in text content -->
  <div class="section">
    <h2>Test 1: Text Content with Store Expressions</h2>
    <p>Auth Status: <span class="status"></span></p>
    <p>User Name: </p>
    <p>User Email: </p>
    <p>Cart Items: </p>
    <p>Cart Total: $</p>
    <p>Current Theme: </p>
  </div>

  <!-- Test 2: Store expressions in attributes (Alpine directives) -->
  <div class="section">
    <h2>Test 2: Store Expressions in Attributes</h2>
    <button @click="$store.auth.login()">Login</button>
    <button @click="$store.auth.logout()">Logout</button>

    <div class="user-info">
      <p><strong>Logged In:</strong> <span></span></p>
      <p><strong>Name:</strong> </p>
      <p><strong>Email:</strong> </p>
    </div>
  </div>

  <!-- Test 3: Cart functionality -->
  <div class="section">
    <h2>Test 3: Cart Store</h2>
    <button @click="$store.cart.addItem('Widget', 9.99)">Add Widget ($9.99)</button>
    <button @click="$store.cart.addItem('Gadget', 19.99)">Add Gadget ($19.99)</button>
    <button @click="$store.cart.addItem('Doohickey', 29.99)">Add Doohickey ($29.99)</button>
    <button @click="$store.cart.clear()">Clear Cart</button>

    <div style="margin-top: 1rem;">
      <p><strong>Items in cart:</strong> </p>
      <p><strong>Total:</strong> $</p>
    </div>
  </div>

  <!-- Test 4: Theme toggle -->
  <div class="section">
    <h2>Test 4: Theme Store</h2>
    <button @click="$store.theme.toggle()">Toggle Theme</button>
    <p>Current theme mode: <strong></strong></p>
  </div>

  <!-- Test 5: Multiple stores in one expression -->
  <div class="section">
    <h2>Test 5: Combined Store Data</h2>
    <p>User  has  items in cart using  theme</p>
  </div>

  <!-- Debug Info -->
  <div class="section" style="background: #fff3cd;">
    <h2>Debug Info</h2>
    <p><strong>What we're testing:</strong></p>
    <ul>
      <li>Store definitions in fence section parse correctly</li>
      <li>Store expressions like  transform to Alpine.js syntax</li>
      <li>Nested properties like  work</li>
      <li>Store methods can be called from Alpine directives</li>
    </ul>
    <p><strong>Expected behavior:</strong></p>
    <ul>
      <li>Initial state shows "false", "Guest", empty cart</li>
      <li>Login button updates auth store and displays user info</li>
      <li>Add item buttons update cart count and total</li>
      <li>Theme toggle switches between "light" and "dark"</li>
    </ul>
  </div>
</body>
</html>`,
  'Header': (props) => `<header class="header">
  <div class="header-container">
    <a href="/" class="brand">
      <svg width="36.206mm" height="7.781mm" version="1.1" viewBox="0 0 36.206 7.781" xmlns="http://www.w3.org/2000/svg">
        <g transform="translate(-264.95 11.104)">
          <path d="m273.59-6.453-2e-3 1.6107c-7e-3 0.22466-0.12785 0.35734-0.35253 0.36241-0.25549 0-0.36051-0.14631-0.36051-0.36123v-4.6353c0-0.092936 0.0232-0.16264 0.0696-0.20911 0.0464-0.052254 0.11906-0.092936 0.2178-0.12198 0.16852-0.052248 0.37175-0.090037 0.6099-0.11327 0.24404-0.023229 0.47924-0.034889 0.70576-0.034889 0.7377 0 1.2751 0.15683 1.612 0.4705 0.34263 0.30786 0.51403 0.72609 0.51403 1.2547 0 0.54021-0.17425 0.97296-0.52275 1.2982-0.34852 0.31948-0.88589 0.47922-1.612 0.47922zm0.83646-0.59249c0.46468 0 0.82484-0.095841 1.0804-0.28753 0.25564-0.19168 0.38343-0.48792 0.38343-0.88872 0-0.39499-0.12489-0.68252-0.37466-0.86259-0.24395-0.18588-0.59546-0.27882-1.0543-0.27882-0.1568 0-0.31363 0.00874-0.47042 0.026165-0.15106 0.011654-0.2876 0.02902-0.40959 0.052251v2.2393z">
          <path d="m290.02-4.9895c0.13651 0 0.23267-0.00934 0.28845-0.027942 0.062-0.024852 0.12724-0.037205 0.19548-0.037205 0.068 0 0.13031 0.027942 0.18614 0.083763 0.0561 0.049626 0.0837 0.12719 0.0837 0.23267 0 0.099268-0.0499 0.17062-0.1489 0.21405-0.21711 0.099268-0.40323 0.14891-0.55839 0.14891-0.15512 0-0.29778-0.00934-0.42812-0.027942-0.12403-0.012426-0.24506-0.052715-0.36289-0.12099-0.27922-0.16132-0.41883-0.44672-0.41883-0.8562v-2.3639h-0.53971c-0.0931 0-0.24964-0.10621-0.25249-0.29294-2e-3 -0.17599 0.15385-0.26609 0.23168-0.26866l0.55806-0.00462 0.0103-0.64389c5e-3 -0.12544 0.16669-0.25549 0.33632-0.25619 0.18493 7.719e-4 0.34881 0.12909 0.3458 0.26617v0.60493h0.86549c0.0871 0 0.15818 0.027942 0.21402 0.083754 0.056 0.055807 0.0837 0.12719 0.0837 0.21405 0 0.086862-0.0279 0.15821-0.0837 0.21405-0.0562 0.055807-0.12713 0.083761-0.21402 0.083761h-0.86549v2.308c0 0.19234 0.0588 0.32263 0.17683 0.39087 0.062 0.037205 0.16125 0.055884 0.29781 0.055884z">
          <g transform="matrix(.77185 0 0 .77185 312.99 -65.312)" fill-rule="evenodd" stroke-width=".26458px">
            <path d="m-54.57 74.523c-1.7919 1.907-5.3906 2.3551-4.3001 4.9881 0.36199 0.74862-0.27419 1.0155-0.50107 0.61951 0 0-3.4086-2.8786-1.4504-4.6156 0.98084-1.0161 6.9801-5.3003 6.3041-1.3029z" fill="#1c7fc7">
            <path d="m-57.629 73.238s-1.5075 0.83503-2.7039 1.8531c0 0 3.0859-1.3869 5.7631-0.56799 0 0 0.39145-1.6984-3.0593-1.2851z" fill="#004a77" fill-opacity=".87434">
            <path d="m-62.252 72.183s0.38056-1.1878 2.1798-1.9508l0.96418 0.28985s6.7226-1.3794 4.5374 4.0016c0 0 0.5771-2.071-4.5695-1.033l-0.70199 0.43123c-1.906-0.26683-2.4099-1.7389-2.4099-1.7389" fill="#22a6ed">
          </g>
          <path d="m277.62-9.6188q0-0.14572 0.10242-0.2368 0.10234-0.10019 0.25131-0.10019 0.14885 0 0.24197 0.10019 0.10234 0.091077 0.10234 0.2368v4.8453q0 0.13662-0.10234 0.2368-0.0931 0.10019-0.24197 0.10019-0.14898 0-0.25131-0.10019-0.10242-0.10018-0.10242-0.2368z">
          <path d="m280.03-6.1675q0.0651 0.51186 0.39083 0.81897 0.33506 0.29781 0.91202 0.29781 0.58636 0 0.99581-0.18613 0.13033-0.065145 0.23266-0.065145 0.10242-0.00934 0.18612 0.074482 0.0931 0.083761 0.0931 0.18613 0 0.20474-0.16752 0.2885-0.15823 0.083758-0.28845 0.14891-0.13034 0.065144-0.27927 0.11168-0.34429 0.093063-0.80033 0.093063-0.93997 0-1.4611-0.52116-0.51184-0.53048-0.51184-1.4984 0-0.83759 0.43737-1.396 0.49331-0.62354 1.3866-0.62354 0.85624 0 1.3495 0.577 0.46533 0.53978 0.46533 1.3494 0 0.1396-0.10234 0.24197-0.0931 0.10237-0.24196 0.10237zm1.126-1.6845q-0.67936 0-0.9865 0.60492-0.11163 0.21405-0.13952 0.51186h2.2521q-0.0277-0.53978-0.4002-0.85621-0.2978-0.26058-0.72589-0.26058z">
          <path d="m285.85-7.7962q-0.67009 0-1.2564 0.73522v2.2801q0 0.1396-0.10244 0.24197-0.10232 0.10237-0.24199 0.10237-0.13951 0-0.24199-0.10237-0.10232-0.10237-0.10232-0.24197v-3.2666q0-0.14891 0.093-0.25128 0.10246-0.10237 0.2513-0.10237 0.13964 0 0.24199 0.10237 0.10244 0.093062 0.10244 0.25128v0.31642q0.30708-0.31642 0.53042-0.45602 0.42814-0.25128 0.81904-0.25128 0.40014 0 0.65139 0.12099 0.25126 0.11167 0.42812 0.32573 0.35368 0.41879 0.35368 1.0423v2.1684q0 0.1396-0.10242 0.24197-0.10234 0.10237-0.24195 0.10237-0.13958 0-0.24199-0.10237-0.10234-0.10237-0.10234-0.24197v-2.094q0-0.42811-0.19544-0.67007-0.19549-0.25127-0.64217-0.25127z">
          <path d="m291.49-8.0475q0-0.15821 0.10242-0.25128 0.10234-0.10237 0.25127-0.10237 0.14887 0 0.24197 0.10237 0.093 0.093062 0.093 0.25128v3.2666q0 0.1396-0.093 0.24197-0.0931 0.10237-0.24197 0.10237-0.14894 0-0.25127-0.10237-0.10242-0.10237-0.10242-0.24197zm0.76312-1.4983q0 0.16751-0.11161 0.27919-0.11174 0.11168-0.27921 0.11168h-0.0465q-0.16751 0-0.27921-0.11168-0.10234-0.11168-0.10234-0.27919v-0.027942q0-0.16752 0.11165-0.2792 0.11167-0.11167 0.26989-0.11167h0.0465q0.16747 0 0.27921 0.11167 0.11161 0.11168 0.11161 0.2792z">
          <path d="m295.26-5.0508q0.42813 0 0.80971-0.20474 0.13025-0.074483 0.24196-0.074483 0.11166-0.00934 0.19538 0.093062 0.0839 0.10237 0.0839 0.24197 0 0.1303-0.17686 0.24198-0.53977 0.35365-1.2098 0.35365-0.84694 0-1.4332-0.53047-0.61426-0.5677-0.60495-1.489 0-0.92135 0.60495-1.489 0.57697-0.53978 1.4332-0.53047 0.45601 0 0.74448 0.12099 0.29781 0.11167 0.46533 0.23266 0.17686 0.11168 0.17686 0.25128 0 0.1396-0.0839 0.23267-0.0837 0.093062-0.16746 0.093062-0.13033 0-0.26988-0.074483-0.39088-0.21405-0.80971-0.20474-0.67006 0-1.0423 0.38156-0.363 0.37226-0.363 0.98649 0 0.61423 0.363 0.99579 0.37224 0.37226 1.0423 0.37226z">
          <path d="m299.21-8.4383q0.88411 0 1.4239 0.55839 0.53971 0.5677 0.52112 1.4611 0 0.90273-0.52112 1.4611-0.53979 0.55839-1.4239 0.55839-0.89347 0-1.4146-0.55839-0.53979-0.55839-0.53979-1.4611 0-0.91204 0.53979-1.4611 0.52111-0.55839 1.4146-0.55839zm-0.85627 3.0991q0.18617 0.15821 0.40954 0.23267 0.23266 0.074482 0.44673 0.074482 0.21402 0 0.43739-0.074482 0.22337-0.074483 0.40943-0.24197 0.4095-0.37226 0.4095-1.0796 0-0.68868-0.4095-1.0702-0.34429-0.31642-0.84682-0.30712-0.82832 0-1.1447 0.74453-0.11171 0.26988-0.11171 0.64215 0 0.37226 0.11171 0.64215 0.11164 0.26989 0.28846 0.43741z">
        </g>
      </svg>
    </a>

    <nav class="nav">
      <a href="/" class="nav-item">Home</a>
      <a href="/products" class="nav-item">Products</a>
      <a href="/login" class="nav-item">Login</a>
      <a href="/register" class="nav-item">Register</a>
    </nav>
  </div>
</header>`,
  'CartBadge': (props) => `<div class="cart-badge">
  <div class="cart-icon-wrapper">
    🛒
    <template x-if="$cart.items.length > 0">
      <span class="cart-count"></span></template>
  </div>
  <div class="cart-details">
    <div class="cart-items"><template x-if="$cart.items.length === 0">
        <span class="empty-cart">Cart is empty</span></template><template x-else>
        <span class="item-count"> items</span></template>
    </div><template x-if="$cart.items.length > 0">
      <div class="cart-total">
        Total: $<strong x-text="$store.cart.formattedTotal"></strong>
      </div></template>
  </div>
</div>`,
  'Dashboard': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>${props.pageTitle} - Custom Go Template</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>

  <main style="padding: 2rem;">
    <div style="background-color: #f0f0f0; padding: 0.5rem 1rem; border-radius: 0.25rem; margin-bottom: 1rem; font-size: 0.875rem; color: #666;">
      ⚡ Build time: <strong>${props.buildTime}</strong>
    </div>

    <h1 style="color: #333; margin-bottom: 1rem;">User Dashboard</h1>
    <p style="margin-bottom: 2rem; color: #666;">Manage your orders, wishlist, and account settings.</p>

    <div style="margin-top: 2rem; padding: 1rem; background-color: #f8f9fa; border-radius: 0.5rem;">
      <p style="margin: 0; color: #666;">
        <a href="/" style="color: #007bff; text-decoration: none;">← Back to Home</a> |
        <a href="/admin" style="color: #007bff; text-decoration: none;">View Admin Dashboard →</a>
      </p>
    </div>
  </main>
</body>
</html>`,
  '../global/header.html': (props) => `<header class="header">
  <div class="header-container">
    <a href="/" class="brand">
      <svg width="36.206mm" height="7.781mm" version="1.1" viewBox="0 0 36.206 7.781" xmlns="http://www.w3.org/2000/svg">
        <g transform="translate(-264.95 11.104)">
          <path d="m273.59-6.453-2e-3 1.6107c-7e-3 0.22466-0.12785 0.35734-0.35253 0.36241-0.25549 0-0.36051-0.14631-0.36051-0.36123v-4.6353c0-0.092936 0.0232-0.16264 0.0696-0.20911 0.0464-0.052254 0.11906-0.092936 0.2178-0.12198 0.16852-0.052248 0.37175-0.090037 0.6099-0.11327 0.24404-0.023229 0.47924-0.034889 0.70576-0.034889 0.7377 0 1.2751 0.15683 1.612 0.4705 0.34263 0.30786 0.51403 0.72609 0.51403 1.2547 0 0.54021-0.17425 0.97296-0.52275 1.2982-0.34852 0.31948-0.88589 0.47922-1.612 0.47922zm0.83646-0.59249c0.46468 0 0.82484-0.095841 1.0804-0.28753 0.25564-0.19168 0.38343-0.48792 0.38343-0.88872 0-0.39499-0.12489-0.68252-0.37466-0.86259-0.24395-0.18588-0.59546-0.27882-1.0543-0.27882-0.1568 0-0.31363 0.00874-0.47042 0.026165-0.15106 0.011654-0.2876 0.02902-0.40959 0.052251v2.2393z">
          <path d="m290.02-4.9895c0.13651 0 0.23267-0.00934 0.28845-0.027942 0.062-0.024852 0.12724-0.037205 0.19548-0.037205 0.068 0 0.13031 0.027942 0.18614 0.083763 0.0561 0.049626 0.0837 0.12719 0.0837 0.23267 0 0.099268-0.0499 0.17062-0.1489 0.21405-0.21711 0.099268-0.40323 0.14891-0.55839 0.14891-0.15512 0-0.29778-0.00934-0.42812-0.027942-0.12403-0.012426-0.24506-0.052715-0.36289-0.12099-0.27922-0.16132-0.41883-0.44672-0.41883-0.8562v-2.3639h-0.53971c-0.0931 0-0.24964-0.10621-0.25249-0.29294-2e-3 -0.17599 0.15385-0.26609 0.23168-0.26866l0.55806-0.00462 0.0103-0.64389c5e-3 -0.12544 0.16669-0.25549 0.33632-0.25619 0.18493 7.719e-4 0.34881 0.12909 0.3458 0.26617v0.60493h0.86549c0.0871 0 0.15818 0.027942 0.21402 0.083754 0.056 0.055807 0.0837 0.12719 0.0837 0.21405 0 0.086862-0.0279 0.15821-0.0837 0.21405-0.0562 0.055807-0.12713 0.083761-0.21402 0.083761h-0.86549v2.308c0 0.19234 0.0588 0.32263 0.17683 0.39087 0.062 0.037205 0.16125 0.055884 0.29781 0.055884z">
          <g transform="matrix(.77185 0 0 .77185 312.99 -65.312)" fill-rule="evenodd" stroke-width=".26458px">
            <path d="m-54.57 74.523c-1.7919 1.907-5.3906 2.3551-4.3001 4.9881 0.36199 0.74862-0.27419 1.0155-0.50107 0.61951 0 0-3.4086-2.8786-1.4504-4.6156 0.98084-1.0161 6.9801-5.3003 6.3041-1.3029z" fill="#1c7fc7">
            <path d="m-57.629 73.238s-1.5075 0.83503-2.7039 1.8531c0 0 3.0859-1.3869 5.7631-0.56799 0 0 0.39145-1.6984-3.0593-1.2851z" fill="#004a77" fill-opacity=".87434">
            <path d="m-62.252 72.183s0.38056-1.1878 2.1798-1.9508l0.96418 0.28985s6.7226-1.3794 4.5374 4.0016c0 0 0.5771-2.071-4.5695-1.033l-0.70199 0.43123c-1.906-0.26683-2.4099-1.7389-2.4099-1.7389" fill="#22a6ed">
          </g>
          <path d="m277.62-9.6188q0-0.14572 0.10242-0.2368 0.10234-0.10019 0.25131-0.10019 0.14885 0 0.24197 0.10019 0.10234 0.091077 0.10234 0.2368v4.8453q0 0.13662-0.10234 0.2368-0.0931 0.10019-0.24197 0.10019-0.14898 0-0.25131-0.10019-0.10242-0.10018-0.10242-0.2368z">
          <path d="m280.03-6.1675q0.0651 0.51186 0.39083 0.81897 0.33506 0.29781 0.91202 0.29781 0.58636 0 0.99581-0.18613 0.13033-0.065145 0.23266-0.065145 0.10242-0.00934 0.18612 0.074482 0.0931 0.083761 0.0931 0.18613 0 0.20474-0.16752 0.2885-0.15823 0.083758-0.28845 0.14891-0.13034 0.065144-0.27927 0.11168-0.34429 0.093063-0.80033 0.093063-0.93997 0-1.4611-0.52116-0.51184-0.53048-0.51184-1.4984 0-0.83759 0.43737-1.396 0.49331-0.62354 1.3866-0.62354 0.85624 0 1.3495 0.577 0.46533 0.53978 0.46533 1.3494 0 0.1396-0.10234 0.24197-0.0931 0.10237-0.24196 0.10237zm1.126-1.6845q-0.67936 0-0.9865 0.60492-0.11163 0.21405-0.13952 0.51186h2.2521q-0.0277-0.53978-0.4002-0.85621-0.2978-0.26058-0.72589-0.26058z">
          <path d="m285.85-7.7962q-0.67009 0-1.2564 0.73522v2.2801q0 0.1396-0.10244 0.24197-0.10232 0.10237-0.24199 0.10237-0.13951 0-0.24199-0.10237-0.10232-0.10237-0.10232-0.24197v-3.2666q0-0.14891 0.093-0.25128 0.10246-0.10237 0.2513-0.10237 0.13964 0 0.24199 0.10237 0.10244 0.093062 0.10244 0.25128v0.31642q0.30708-0.31642 0.53042-0.45602 0.42814-0.25128 0.81904-0.25128 0.40014 0 0.65139 0.12099 0.25126 0.11167 0.42812 0.32573 0.35368 0.41879 0.35368 1.0423v2.1684q0 0.1396-0.10242 0.24197-0.10234 0.10237-0.24195 0.10237-0.13958 0-0.24199-0.10237-0.10234-0.10237-0.10234-0.24197v-2.094q0-0.42811-0.19544-0.67007-0.19549-0.25127-0.64217-0.25127z">
          <path d="m291.49-8.0475q0-0.15821 0.10242-0.25128 0.10234-0.10237 0.25127-0.10237 0.14887 0 0.24197 0.10237 0.093 0.093062 0.093 0.25128v3.2666q0 0.1396-0.093 0.24197-0.0931 0.10237-0.24197 0.10237-0.14894 0-0.25127-0.10237-0.10242-0.10237-0.10242-0.24197zm0.76312-1.4983q0 0.16751-0.11161 0.27919-0.11174 0.11168-0.27921 0.11168h-0.0465q-0.16751 0-0.27921-0.11168-0.10234-0.11168-0.10234-0.27919v-0.027942q0-0.16752 0.11165-0.2792 0.11167-0.11167 0.26989-0.11167h0.0465q0.16747 0 0.27921 0.11167 0.11161 0.11168 0.11161 0.2792z">
          <path d="m295.26-5.0508q0.42813 0 0.80971-0.20474 0.13025-0.074483 0.24196-0.074483 0.11166-0.00934 0.19538 0.093062 0.0839 0.10237 0.0839 0.24197 0 0.1303-0.17686 0.24198-0.53977 0.35365-1.2098 0.35365-0.84694 0-1.4332-0.53047-0.61426-0.5677-0.60495-1.489 0-0.92135 0.60495-1.489 0.57697-0.53978 1.4332-0.53047 0.45601 0 0.74448 0.12099 0.29781 0.11167 0.46533 0.23266 0.17686 0.11168 0.17686 0.25128 0 0.1396-0.0839 0.23267-0.0837 0.093062-0.16746 0.093062-0.13033 0-0.26988-0.074483-0.39088-0.21405-0.80971-0.20474-0.67006 0-1.0423 0.38156-0.363 0.37226-0.363 0.98649 0 0.61423 0.363 0.99579 0.37224 0.37226 1.0423 0.37226z">
          <path d="m299.21-8.4383q0.88411 0 1.4239 0.55839 0.53971 0.5677 0.52112 1.4611 0 0.90273-0.52112 1.4611-0.53979 0.55839-1.4239 0.55839-0.89347 0-1.4146-0.55839-0.53979-0.55839-0.53979-1.4611 0-0.91204 0.53979-1.4611 0.52111-0.55839 1.4146-0.55839zm-0.85627 3.0991q0.18617 0.15821 0.40954 0.23267 0.23266 0.074482 0.44673 0.074482 0.21402 0 0.43739-0.074482 0.22337-0.074483 0.40943-0.24197 0.4095-0.37226 0.4095-1.0796 0-0.68868-0.4095-1.0702-0.34429-0.31642-0.84682-0.30712-0.82832 0-1.1447 0.74453-0.11171 0.26988-0.11171 0.64215 0 0.37226 0.11171 0.64215 0.11164 0.26989 0.28846 0.43741z">
        </g>
      </svg>
    </a>

    <nav class="nav">
      <a href="/" class="nav-item">Home</a>
      <a href="/products" class="nav-item">Products</a>
      <a href="/login" class="nav-item">Login</a>
      <a href="/register" class="nav-item">Register</a>
    </nav>
  </div>
</header>`,
  'Store-test-with-theme': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>${props.title}</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body :style="$store.theme.bodyStyle">

  <h1 style="text-align: center;">${props.title}</h1>
  <p style="text-align: center;">Current theme: <strong></strong></p>

  <!-- APPROACH 1: Inline Styles in Store (using getters) -->
  <div :style="$store.theme.sectionStyle">
    <h2 style="margin-top: 0;">🎨 Approach 1: Inline Styles in Store (Getters)</h2>
    <p><strong>Pattern:</strong> Store defines <code>get bodyStyle()</code>, <code>get sectionStyle()</code>, <code>get buttonStyle()</code></p>
    <p><strong>Usage:</strong> <code>:style="$store.theme.buttonStyle"</code></p>
    <button @click="$store.theme.toggle()" :style="$store.theme.buttonStyle">
      Toggle to  Mode
    </button>
    <p style="margin-top: 1rem;">✅ <strong>Benefits:</strong> Single source of truth, DRY, all theme logic in the store</p>
    <p>✅ <strong>Clean HTML:</strong> No duplication of dark/light values in templates</p>
  </div>

  <!-- APPROACH 2: CSS Classes -->
  <div class="theme-section" :class="$store.theme.mode">
    <h2 style="margin-top: 0;">🎨 Approach 2: CSS Classes with :class Binding</h2>
    <p><strong>Pattern:</strong> CSS defines <code>.theme-section.light</code> and <code>.theme-section.dark</code></p>
    <p><strong>Usage:</strong> <code>:class="$store.theme.mode"</code></p>
    <button class="theme-btn" :class="$store.theme.mode" @click="$store.theme.toggle()">
      Toggle to  Mode
    </button>
    <p style="margin-top: 1rem;">✅ <strong>Benefits:</strong> Separation of concerns, CSS in stylesheets, better performance</p>
    <p>✅ <strong>Production Ready:</strong> Standard CSS approach, easier for designers</p>
  </div>

  <!-- Auth Section -->
  <div :style="$store.theme.mode === 'dark' ? 'background: #2d2d2d; padding: 1.5rem; margin-bottom: 1rem; border-radius: 8px;' : 'background: #f5f5f5; padding: 1.5rem; margin-bottom: 1rem; border-radius: 8px;'">
    <h2 style="margin-top: 0;">✅ Authentication Store</h2>
    <p>Status: <strong></strong></p>
    <p>User: <strong></strong></p>
    <p>Email: <strong></strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.auth.login()" :style="$store.theme.mode === 'dark' ? 'background: #4a9eff; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem;' : 'background: #2563eb; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem;'">
        Login
      </button>
      <button @click="$store.auth.logout()" :style="$store.theme.mode === 'dark' ? 'background: #666; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer;' : 'background: #6b7280; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer;'">
        Logout
      </button>
    </div>
  </div>

  <!-- Cart Section -->
  <div :style="$store.theme.mode === 'dark' ? 'background: #2d2d2d; padding: 1.5rem; margin-bottom: 1rem; border-radius: 8px;' : 'background: #f5f5f5; padding: 1.5rem; margin-bottom: 1rem; border-radius: 8px;'">
    <h2 style="margin-top: 0;">🛒 Shopping Cart Store</h2>
    <p>Items in cart: <strong></strong></p>
    <p>Total: <strong>$</strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.cart.addItem('Widget', 9.99)" :style="$store.theme.mode === 'dark' ? 'background: #4a9eff; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;' : 'background: #2563eb; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;'">
        Add Widget ($9.99)
      </button>
      <button @click="$store.cart.addItem('Gadget', 19.99)" :style="$store.theme.mode === 'dark' ? 'background: #4a9eff; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;' : 'background: #2563eb; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;'">
        Add Gadget ($19.99)
      </button>
      <button @click="$store.cart.addItem('Doohickey', 29.99)" :style="$store.theme.mode === 'dark' ? 'background: #4a9eff; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;' : 'background: #2563eb; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-right: 0.5rem; margin-bottom: 0.5rem;'">
        Add Doohickey ($29.99)
      </button>
      <button @click="$store.cart.clear()" :style="$store.theme.mode === 'dark' ? 'background: #dc2626; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer;' : 'background: #dc2626; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer;'">
        Clear Cart
      </button>
    </div>
  </div>

  <!-- Comparison Section -->
  <div :style="$store.theme.sectionStyle">
    <h2 style="margin-top: 0;">📊 Approach Comparison</h2>
    <table style="width: 100%; border-collapse: collapse;">
      <thead>
        <tr>
          <th style="text-align: left; padding: 0.5rem; border-bottom: 2px solid #ccc;">Aspect</th>
          <th style="text-align: left; padding: 0.5rem; border-bottom: 2px solid #ccc;">Approach 1: Store Getters</th>
          <th style="text-align: left; padding: 0.5rem; border-bottom: 2px solid #ccc;">Approach 2: CSS Classes</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>Code Location</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">Styles in store (fence)</td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">Styles in <code>&lt;style&gt;</code> tag</td>
        </tr>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>HTML Usage</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><code>:style="$store.theme.buttonStyle"</code></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><code>:class="$store.theme.mode"</code></td>
        </tr>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>DRY Principle</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">✅ Excellent - single source</td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">✅ Excellent - CSS only</td>
        </tr>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>Performance</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">Good - inline styles</td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">✅ Better - CSS caching</td>
        </tr>
        <tr>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;"><strong>Designer Friendly</strong></td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">JavaScript knowledge needed</td>
          <td style="padding: 0.5rem; border-bottom: 1px solid #eee;">✅ Standard CSS workflow</td>
        </tr>
        <tr>
          <td style="padding: 0.5rem;"><strong>Best For</strong></td>
          <td style="padding: 0.5rem;">Dynamic, complex theming logic</td>
          <td style="padding: 0.5rem;">✅ Production apps, team workflows</td>
        </tr>
      </tbody>
    </table>
  </div>

</body>
</html>`,
  '../global/footer.html': (props) => `<!-- ============================================ --><!--                   Footer                     --><!-- ============================================ --><footer id="cs-footer-2440">
    <div class="cs-container">
        <div class="cs-top">
            <div class="cs-logo-group">
                <a class="cs-logo" aria-label="go back to home" href="">
                    <img class="cs-logo-img" aria-hidden="true" loading="lazy" decoding="async" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/artistitch.svg" alt="logo" width="260" height="48"></img>
                </a>
                <ul class="cs-contact-group">
                    <li class="cs-contact-li">
                        <img class="cs-contact-icon" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/white-phone.svg" alt="phone icon" height="24" width="24" loading="lazy" decoding="async" aria-hidden="true"></img>
                        <a class="cs-contact-link" href="tel:555-474-7386">555-474-7386</a>
                    </li>
                    <li class="cs-contact-li">
                        <img class="cs-contact-icon" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/white-email.svg" alt="email icon" height="24" width="24" loading="lazy" decoding="async" aria-hidden="true"></img>
                        <a class="cs-contact-link" href="mailto:ArtiStitchart@gmail.com">ArtiStitchart@gmail.com</a>
                    </li>
                    <li class="cs-contact-li">
                        <img class="cs-contact-icon" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/white-location.svg" alt="location icon" height="24" width="24" loading="lazy" decoding="async" aria-hidden="true"></img>
                        <a class="cs-contact-link" href="">555 Fake St, Portland, OR 97217, United States</a>
                    </li>
                </ul>
            </div>
            <!--Sitemap-->
            <ul class="cs-ul">
                <li class="cs-li">
                    <span class="cs-header">Quick Links</span>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Home</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">About</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Services</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Artist</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">FAQ</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Contact</a>
                </li>
            </ul>
            <ul class="cs-ul">
                <li class="cs-li">
                    <span class="cs-header">Our Services</span>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Custom UV Body Painting</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Spiritual & Symbolic Art</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Workshops & Education</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Body Art Reworks</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Live Event Painting</a>
                </li>
                <li class="cs-li">
                    <a class="cs-link" href="">Symbolic Expression Art</a>
                </li>
            </ul>
            <div class="cs-subscribe-group">
                <span class="cs-header">Sign up to receive the latest news and events from us.</span>
                <form class="cs-form" name="Footer Subscribe" method="post">
                    <input class="cs-input" aria-label="email" name="email" type="email" placeholder="Email Address" required=""></input>
                    <button class="cs-button-solid" name="submit" type="submit" aria-label="submit">
                        <img class="cs-icon" src="https://nyc3.digitaloceanspaces.com/csimages2/Icons/white-upper-right-arrow.svg" alt="arrow icon" height="24" width="24" loading="lazy" decoding="async" aria-hidden="true"></img>
                    </button>
                </form>
            </div>
        </div>
        <!--Bottom Copyright And Social-->
        <div class="cs-bottom">
            <div class="cs-flex">
                <a href="" class="cs-link">Facebook</a>
                <div class="cs-divider" aria-hidden="true"></div>
                <a href="" class="cs-link">Instagram</a>
                <div class="cs-divider" aria-hidden="true"></div>
                <a href="" class="cs-link">Tiktok</a>
                <div class="cs-divider" aria-hidden="true"></div>
                <a href="" class="cs-link">Youtube</a>
            </div>
            <span class="cs-copyright-group">
                <span class="cs-copyright">© Copyright 2025 -</span>
                <a class="cs-link" href="">ArtiStitch</a>
            </span>
        </div>
    </div>
</footer>`,
  'Admin': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>${props.pageTitle} - Custom Go Template</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>

  <main style="padding: 2rem;">
    <div style="background-color: #f0f0f0; padding: 0.5rem 1rem; border-radius: 0.25rem; margin-bottom: 1rem; font-size: 0.875rem; color: #666;">
      ⚡ Build time: <strong>${props.buildTime}</strong>
    </div>

    <h1 style="color: #333; margin-bottom: 1rem;">Admin Dashboard</h1>
    <p style="margin-bottom: 2rem; color: #666;">Welcome to the admin control panel. Monitor site statistics and recent activity.</p>

    <div style="margin-top: 2rem; padding: 1rem; background-color: #f8f9fa; border-radius: 0.5rem;">
      <p style="margin: 0; color: #666;">
        <a href="/" style="color: #007bff; text-decoration: none;">← Back to Home</a> |
        <a href="/user-dashboard" style="color: #007bff; text-decoration: none;">View User Dashboard →</a>
      </p>
    </div>
  </main>
</body>
</html>`,
  '../components/adminpanel.html': (props) => `<div class="admin-panel">
  <div class="admin-header">
    <div class="admin-title">Admin Dashboard</div>
    <div class="admin-user">
      <span>${props.user.name}</span>
      <span class="admin-badge">${props.user.role}</span>
    </div>
  </div>
  
  <div class="admin-content">
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-value">${props.stats.users}</div>
        <div class="stat-label">Registered Users</div>
      </div>
      
      <div class="stat-card">
        <div class="stat-value">${props.stats.products}</div>
        <div class="stat-label">Products</div>
      </div>
      
      <div class="stat-card">
        <div class="stat-value">${props.stats.orders}</div>
        <div class="stat-label">Orders</div>
      </div>
      
      <div class="stat-card">
        <div class="stat-value">${props.formatCurrency(stats.revenue)}</div>
        <div class="stat-label">Total Revenue</div>
      </div>
    </div>
    
    <h3 class="section-title">Recent Activity</h3>
    <ul class="action-list"><template x-for="action,  in recentActions">
        <li class="action-item">
          <div class="action-info">
            <span class="action-desc">${props.action.action}</span>
            <span class="action-user">by ${props.action.user}</span>
          </div>
          <div class="action-time">${props.formatTimestamp(action.timestamp)}</div>
        </li></template>
    </ul>
    
    <div class="admin-actions">
      <button class="admin-button">Add New Product</button>
      <button class="admin-button">Manage Users</button>
      <button class="admin-button admin-button-secondary">View All Activity</button>
    </div>
  </div>
</div>`,
  '../components/age.html': (props) => `<div class="age-container">
  <div class="age-info">${props.name}'s age is: ${props.age}</div><template x-if="age > 300">
    <div class="age-badge dust">Dust (300+)</div></template><template x-else-if="age > 200">
    <div class="age-badge dead">Dead (200+)</div></template><template x-else-if="age > 90">
    <div class="age-badge old">Wow (90+)</div></template><template x-else-if="age > 80">
    <div class="age-badge old">Well done (80+)</div></template><template x-else-if="age > 40">
    <div class="age-badge mid">Over the hill (40+)</div></template><template x-else-if="age >= 21">
    <div class="age-badge young">Can drink (21+)</div></template><template x-else>
    <div class="age-badge new">Just a baby (Under 21)</div></template>
</div>`,
  'Productcard': (props) => `<div class="product-card {featured ? 'featured' : ''}">
  <div class="product-name">
    ${props.name}<template x-if="featured">
      <span class="featured-badge">Featured</span></template>
  </div>

  <div class="product-price">$${props.price}</div><template x-if="inStock">
    <div class="product-status in-stock">In Stock</div></template><template x-else>
    <div class="product-status out-of-stock">Out of Stock</div></template><template x-if="tags && tags.length > 0">
    <div class="product-tags"><template x-for="tag,  in tags">
        <span class="product-tag">${props.tag}</span></template>
    </div></template><template x-if="inStock">
    <button class="add-to-cart">Add to Cart</button></template><template x-else>
    <button class="add-to-cart" disabled="">Sold Out</button></template>
</div>`,
  '../components/productcard.html': (props) => `<div class="product-card {featured ? 'featured' : ''}">
  <div class="product-name">
    ${props.name}<template x-if="featured">
      <span class="featured-badge">Featured</span></template>
  </div>

  <div class="product-price">$${props.price}</div><template x-if="inStock">
    <div class="product-status in-stock">In Stock</div></template><template x-else>
    <div class="product-status out-of-stock">Out of Stock</div></template><template x-if="tags && tags.length > 0">
    <div class="product-tags"><template x-for="tag,  in tags">
        <span class="product-tag">${props.tag}</span></template>
    </div></template><template x-if="inStock">
    <button class="add-to-cart">Add to Cart</button></template><template x-else>
    <button class="add-to-cart" disabled="">Sold Out</button></template>
</div>`,
  'LoginStatus': (props) => `<div class="login-status"><template x-if="$auth.isLoggedIn">
    <div class="logged-in">
      <span class="user-info">
        👤 Welcome, <strong></strong>
      </span>
      <button @click="$store.auth.logout()" class="btn btn-secondary">
        Logout
      </button>
    </div></template><template x-else>
    <div class="logged-out">
      <span class="guest-info">👋 Guest</span>
      <button @click="$store.auth.login()" class="btn btn-primary">
        Login
      </button>
    </div></template>
</div>`,
  '../components/LoginStatus.html': (props) => `<div class="login-status"><template x-if="$auth.isLoggedIn">
    <div class="logged-in">
      <span class="user-info">
        👤 Welcome, <strong></strong>
      </span>
      <button @click="$store.auth.logout()" class="btn btn-secondary">
        Logout
      </button>
    </div></template><template x-else>
    <div class="logged-out">
      <span class="guest-info">👋 Guest</span>
      <button @click="$store.auth.login()" class="btn btn-primary">
        Login
      </button>
    </div></template>
</div>`,
  'Notification': (props) => `<div class="notification notification-${props.type}"><template x-if="type == "info"">
    <div class="notification-icon">ℹ️</div></template><template x-else-if="type == "success"">
    <div class="notification-icon">✓</div></template><template x-else-if="type == "warning"">
    <div class="notification-icon">⚠️</div></template><template x-else>
    <div class="notification-icon">✕</div></template>
  <div class="notification-message">${props.message}</div><template x-if="dismissable">
    <button class="notification-dismiss" @click="currentNotification = null">×</button></template>
</div>`,
  '../components/notification.html': (props) => `<div class="notification notification-${props.type}"><template x-if="type == "info"">
    <div class="notification-icon">ℹ️</div></template><template x-else-if="type == "success"">
    <div class="notification-icon">✓</div></template><template x-else-if="type == "warning"">
    <div class="notification-icon">⚠️</div></template><template x-else>
    <div class="notification-icon">✕</div></template>
  <div class="notification-message">${props.message}</div><template x-if="dismissable">
    <button class="notification-dismiss" @click="currentNotification = null">×</button></template>
</div>`,
  'Head': (props) => `<head>
	<meta charset="UTF-8"></meta>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
	<title>${props.title}</title>
	<meta name="description" content="${props.description}"></meta>

	<!-- Alpine.js - Global reactive framework (MUST load first) -->
	<script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>

	<!-- Runtime Components - Load as module to ensure proper timing -->
	<!-- CRITICAL: Must load BEFORE Alpine initializes, so use immediate execution -->
	<script src="/static/js/runtime-components.js"></script>

	<!-- App Scripts -->
	<script defer="" type="text/javascript" src="/scripts/script.js"></script>
	<script defer="" type="text/javascript" src="/scripts/cms.js"></script>

	<!-- App Styles -->
	<link rel="stylesheet" href="/styles/style.css"></link>
	<link rel="stylesheet" href="/styles/cms.css"></link>
	  <link rel="preload" href="/static/fonts/roboto-v29-latin-regular.woff2 " as="font" type="font/woff2" crossorigin="">
</head>`,
  '../content/admin.html': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>${props.pageTitle} - Custom Go Template</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>

  <main style="padding: 2rem;">
    <div style="background-color: #f0f0f0; padding: 0.5rem 1rem; border-radius: 0.25rem; margin-bottom: 1rem; font-size: 0.875rem; color: #666;">
      ⚡ Build time: <strong>${props.buildTime}</strong>
    </div>

    <h1 style="color: #333; margin-bottom: 1rem;">Admin Dashboard</h1>
    <p style="margin-bottom: 2rem; color: #666;">Welcome to the admin control panel. Monitor site statistics and recent activity.</p>

    <div style="margin-top: 2rem; padding: 1rem; background-color: #f8f9fa; border-radius: 0.5rem;">
      <p style="margin: 0; color: #666;">
        <a href="/" style="color: #007bff; text-decoration: none;">← Back to Home</a> |
        <a href="/user-dashboard" style="color: #007bff; text-decoration: none;">View User Dashboard →</a>
      </p>
    </div>
  </main>
</body>
</html>`,
  '../content/store-test-minimal.html': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <title>${props.title}</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css"></link>
</head>
<body style="padding: 2rem;">
  <main class="container">
    <h1 style="text-align: center; color: #2563eb;">${props.title}</h1>

    <article>
    <h2>✅ Test 1: Authentication Store</h2>
    <p>Status: <span class="status"></span></p>
    <p>User: <strong></strong></p>
    <p>Email: <strong></strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.auth.login()">Login</button>
      <button @click="$store.auth.logout()">Logout</button>
    </div>

    <div class="info-box">
      <strong>Try it:</strong> Click "Login" and watch the values update reactively!<br></br>
      Store expressions: <code></code>, <code></code>
    </div>
  </div>

  <div class="test-section">
    <h2>🛒 Test 2: Shopping Cart Store</h2>
    <p>Items in cart: <strong></strong></p>
    <p>Total: <strong>$</strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.cart.addItem('Widget', 9.99)">Add Widget ($9.99)</button>
      <button @click="$store.cart.addItem('Gadget', 19.99)">Add Gadget ($19.99)</button>
      <button @click="$store.cart.addItem('Doohickey', 29.99)">Add Doohickey ($29.99)</button>
      <button @click="$store.cart.clear()" class="secondary">Clear Cart</button>
    </div>

    <div class="info-box">
      <strong>Try it:</strong> Add items and watch the count and total update!<br></br>
      Store expressions: <code></code>, <code></code>
    </div>
  </div>

  <div class="test-section">
    <h2>🎨 Test 3: Theme Store</h2>
    <p>Current theme: <strong></strong></p>

    <div style="margin-top: 1rem;">
      <button @click="$store.theme.toggle()">Toggle Theme</button>
    </div>

    <div class="info-box">
      <strong>Try it:</strong> Click to toggle between "light" and "dark"<br></br>
      Store expression: <code></code>
    </div>
  </div>

  div class="test-section">
    <h2>🎉 Store System Features Working</h2>
    <ul>
      <li>✅ <strong>Inline store definitions</strong> in fence section</li>
      <li>✅ <strong>Store expressions</strong> transform to Alpine.js syntax</li>
      <li>✅ <strong>Store initialization</strong> via <code>Alpine.store()</code></li>
      <li>✅ <strong>Reactive updates</strong> across the page</li>
      <li>✅ <strong>Store methods</strong> callable from Alpine directives</li>
      <li>✅ <strong>Nested properties</strong> like <code>$auth.user.name</code></li>
    </ul>

    <div class="info-box">
      <strong>View Page Source</strong> to see:<br></br>
      • Store initialization script in <code>&lt;head&gt;</code><br></br>
      • Store expressions transformed to <code>&lt;span x-text="$store.X.Y"&gt;</code><br></br>
      • Alpine.js event handlers with <code>@click="$store.X.method()"</code>
      </div>
    </article>
  </main>
</body>
</html>`,
  'Todos': (props) => `<table class="todos-table">
  <thead>
    <tr>
      <th style="width: 40px; text-align: center;">#</th>
      <th>Task</th>
      <th style="width: 120px; text-align: center;">Status</th>
    </tr>
  </thead>
  <tbody><template x-for="todo, index in todos.slice(start, start + number)">
      <tr>
        <td style="text-align: center; color: #6b7280; font-weight: 500;">${(props.start * 1) + index + 1}</td>
        <td style="{todo.completed ? 'text-decoration: line-through; color: #9ca3af;' : ''}">${todo.title}</td>
        <td style="text-align: center;">
          <span style="{todo.completed ? 'color: #16a34a; font-weight: 600;' : 'color: #dc2626; font-weight: 600;'}">
            ${todo.completed ? props.'✓ props.Done' : props.'○ props.Pending'}
          </span>
        </td>
      </tr></template>
  </tbody>
</table>`,
  '../content/simple.html': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <title>Basic Test</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
  <h1>Basic Template Test</h1>

  <div>
    <h2>Expressions</h2>
    <p>Name: ${props.name}</p>
    <p>Age: ${props.age}</p>
  </div>

  <div>
    <h2>Conditionals</h2><template x-if="age >= 18">
      <p>${props.name} is an adult</p></template><template x-else>
      <p>${props.name} is a minor</p></template>
  </div>

  <div>
    <h2>Loops</h2>
    <ul><template x-for="item,  in items">
        <li>${item}</li></template>
    </ul>
  </div>
</body>
</html>`,
  'Adminpanel': (props) => `<div class="admin-panel">
  <div class="admin-header">
    <div class="admin-title">Admin Dashboard</div>
    <div class="admin-user">
      <span>${props.user.name}</span>
      <span class="admin-badge">${props.user.role}</span>
    </div>
  </div>
  
  <div class="admin-content">
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-value">${props.stats.users}</div>
        <div class="stat-label">Registered Users</div>
      </div>
      
      <div class="stat-card">
        <div class="stat-value">${props.stats.products}</div>
        <div class="stat-label">Products</div>
      </div>
      
      <div class="stat-card">
        <div class="stat-value">${props.stats.orders}</div>
        <div class="stat-label">Orders</div>
      </div>
      
      <div class="stat-card">
        <div class="stat-value">${props.formatCurrency(stats.revenue)}</div>
        <div class="stat-label">Total Revenue</div>
      </div>
    </div>
    
    <h3 class="section-title">Recent Activity</h3>
    <ul class="action-list"><template x-for="action,  in recentActions">
        <li class="action-item">
          <div class="action-info">
            <span class="action-desc">${props.action.action}</span>
            <span class="action-user">by ${props.action.user}</span>
          </div>
          <div class="action-time">${props.formatTimestamp(action.timestamp)}</div>
        </li></template>
    </ul>
    
    <div class="admin-actions">
      <button class="admin-button">Add New Product</button>
      <button class="admin-button">Manage Users</button>
      <button class="admin-button admin-button-secondary">View All Activity</button>
    </div>
  </div>
</div>`,
  '../components/todos.html': (props) => `<table class="todos-table">
  <thead>
    <tr>
      <th style="width: 40px; text-align: center;">#</th>
      <th>Task</th>
      <th style="width: 120px; text-align: center;">Status</th>
    </tr>
  </thead>
  <tbody><template x-for="todo, index in todos.slice(start, start + number)">
      <tr>
        <td style="text-align: center; color: #6b7280; font-weight: 500;">${(props.start * 1) + index + 1}</td>
        <td style="{todo.completed ? 'text-decoration: line-through; color: #9ca3af;' : ''}">${todo.title}</td>
        <td style="text-align: center;">
          <span style="{todo.completed ? 'color: #16a34a; font-weight: 600;' : 'color: #dc2626; font-weight: 600;'}">
            ${todo.completed ? props.'✓ props.Done' : props.'○ props.Pending'}
          </span>
        </td>
      </tr></template>
  </tbody>
</table>`,
  '../global/html.html': (props) => `<html lang="en">

<!-- Global Head Component -->

<body>
	<!-- User Menu (conditional - if authenticated) -->
	<!-- {#if user && $user.isAuthenticated}
		<UserMenu user={user} content={content} shadowContent={shadowContent} />
	{/if} -->

	<!-- Login Modal (conditional) -->
	<!-- <LoginModal /> -->

	<main>
		<!-- Global Navigation -->

		<!-- Global Header -->
		<!-- <Header /> -->

		<!-- ============================================ -->
		<!-- DYNAMIC CONTENT LAYOUT INJECTION             -->
		<!-- This is where page-specific layouts render  -->
		<!-- For pages â renders layouts/content/pages.html -->
		<!-- For courses â renders layouts/content/courses.html -->
		<!-- ============================================ -->

		<!-- Dynamic layout component injection -->
		<!-- In Svelte this is: <svelte:component this={layout} {...content.fields} /> -->
		<!-- In our Go template engine, this is handled by the server -->
		<!-- The server will inject the appropriate layout based on the route -->

		<!-- Slot for dynamic content - replaced by server with specific layout -->
		<!-- CRITICAL FIX: Pass content and allContent explicitly to layout component -->
		<!-- Pages.html needs content.components for iteration -->
		<!-- Newsletter Subscription (optional global component) -->
		<!-- <Newsletter /> -->

		<!-- Global Footer -->
	</main>

</body>

</html>`,
  '../content/_index.html': (props) => ``,
  '../content/pages.html': (props) => `<!-- Pages Layout - Renders dynamic components from content JSON --><!-- NOTE: <head> provided by wrapper layout (includes Alpine.js + runtime-components.js) --><!-- NOTE: No fence section - uses variables from parent scope (html.html wrapper) --><template x-for="component,  in content.components"></template>`,
  'Simple': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <title>Basic Test</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
  <h1>Basic Template Test</h1>

  <div>
    <h2>Expressions</h2>
    <p>Name: ${props.name}</p>
    <p>Age: ${props.age}</p>
  </div>

  <div>
    <h2>Conditionals</h2><template x-if="age >= 18">
      <p>${props.name} is an adult</p></template><template x-else>
      <p>${props.name} is a minor</p></template>
  </div>

  <div>
    <h2>Loops</h2>
    <ul><template x-for="item,  in items">
        <li>${item}</li></template>
    </ul>
  </div>
</body>
</html>`,
  '../components/userprofile.html': (props) => `<template x-if="compact">
  <div class="profile-card profile-compact">
    <div class="profile-avatar-small"><template x-if="user.avatar">
        <img src="${props.user.avatar}" alt="${props.user.name}"></template><template x-else-if="user.name">
        ${props.user.name.charAt(0)}</template><template x-else>
        U</template>
    </div>
    <div>
      <div class="profile-name-compact">${props.user.name}</div><template x-if="showRole">
        <span class="profile-role ${props.roleBadge.bg} ${props.roleBadge.text}">
          ${props.user.role}
        </span></template>
    </div>
  </div></template><template x-else>
  <div class="profile-card">
    <div class="profile-header">
      <div class="profile-avatar"><template x-if="user.avatar">
          <img src="${props.user.avatar}" alt="${props.user.name}"></template><template x-else-if="user.name">
          ${props.user.name.charAt(0)}</template><template x-else>
          U</template>
      </div>
      <h2 class="profile-name">${props.user.name}</h2><template x-if="showRole">
        <span class="profile-role ${props.roleBadge.bg} ${props.roleBadge.text}">
          ${props.user.role}
        </span></template>
    </div>
    <div class="profile-details"><template x-if="showEmail">
        <div class="profile-detail">
          <div class="profile-label">Email</div>
          <div class="profile-value">${props.user.email}</div>
        </div></template><template x-if="showJoinDate">
        <div class="profile-detail">
          <div class="profile-label">Member since</div>
          <div class="profile-value">${props.formattedJoinDate}</div>
        </div></template>
    </div>
  </div></template>`,
  '../content/dashboard.html': (props) => `<html lang="en">
<head>
  <meta charset="UTF-8"></meta>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
  <title>${props.pageTitle} - Custom Go Template</title>
  <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>

  <main style="padding: 2rem;">
    <div style="background-color: #f0f0f0; padding: 0.5rem 1rem; border-radius: 0.25rem; margin-bottom: 1rem; font-size: 0.875rem; color: #666;">
      ⚡ Build time: <strong>${props.buildTime}</strong>
    </div>

    <h1 style="color: #333; margin-bottom: 1rem;">User Dashboard</h1>
    <p style="margin-bottom: 2rem; color: #666;">Manage your orders, wishlist, and account settings.</p>

    <div style="margin-top: 2rem; padding: 1rem; background-color: #f8f9fa; border-radius: 0.5rem;">
      <p style="margin: 0; color: #666;">
        <a href="/" style="color: #007bff; text-decoration: none;">← Back to Home</a> |
        <a href="/admin" style="color: #007bff; text-decoration: none;">View Admin Dashboard →</a>
      </p>
    </div>
  </main>
</body>
</html>`,
  '../content/store-demo.html': (props) => `<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Global Store System - Component Demo</title>
    <script defer="" src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
  </head>
  <body :style="\`background-color: $${$store.theme.getCurrentColors().background}; color: $${$store.theme.getCurrentColors().text}; --code-bg: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#5a7ec5'}; transition: all 0.3s ease;\`">
    <div class="container">
      <h1 :style="\`color: $${$store.theme.getCurrentColors().text};\`">
        Global Store System - Component Demo
      </h1>
      <p class="subtitle" :style="\`color: $${$store.theme.getCurrentColors().secondary};\`">
        Reusable components demonstrating the Global Store System
      </p>

      <!-- Header with Store Components -->
      <div class="header" :style="\`background: ${$store.theme.mode === 'light' ? 'white' : '#2d3748'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
      </div>

      <!-- Interactive Demo Section -->
      <div class="demo-section" :style="\`background: ${$store.theme.mode === 'light' ? 'white' : '#2d3748'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
        <h2>🧪 Try the Interactive Demo</h2>
        <p>
          These buttons modify the stores, and all components above update
          automatically:
        </p>

        <!-- Auth Actions -->
        <div class="action-group" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>Authentication Actions</h3>
          <button @click="$store.auth.login()" class="btn btn-primary">
            Login as Test User
          </button>
          <button @click="$store.auth.logout()" class="btn btn-secondary">
            Logout
          </button>
        </div>

        <!-- Cart Actions -->
        <div class="action-group" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>Shopping Cart Actions</h3>
          <button @click="$store.cart.addItem({ name: 'Widget', price: 9.99 })" class="btn btn-success">
            Add Widget ($9.99)
          </button>
          <button @click="$store.cart.addItem({ name: 'Gadget', price: 19.99 })" class="btn btn-success">
            Add Gadget ($19.99)
          </button>
          <button @click="$store.cart.addItem({ name: 'Doohickey', price: 29.99 })" class="btn btn-success">
            Add Doohickey ($29.99)
          </button>
          <button @click="$store.cart.clear()" class="btn btn-danger">
            Clear Cart
          </button>
        </div>

        <!-- Theme Actions -->
        <div class="action-group" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>Theme Actions</h3>
          <p class="note">
            The ThemeToggle component above provides its own buttons, but you
            can also control the theme from anywhere in your app:
          </p>
          <button @click="$store.theme.setLight()" class="btn btn-light">
            Set Light Mode
          </button>
          <button @click="$store.theme.setDark()" class="btn btn-dark">
            Set Dark Mode
          </button>
          <div style="
              margin-top: 1rem;
              padding: 1rem;
              background: rgba(59, 130, 246, 0.1);
              border-radius: 6px;
              border-left: 4px solid #3b82f6;
            ">
            <strong>Visual Feedback Test:</strong>
            <p style="margin-top: 0.5rem; font-size: 0.95rem">
              Click the theme buttons above and watch the entire page change
              colors! Current mode: <strong x-text="$store.theme.mode"></strong>
            </p>
          </div>
        </div>
      </div>

      <!-- Store State Display -->
      <div class="state-section" :style="\`background: ${$store.theme.mode === 'light' ? 'white' : '#2d3748'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
        <h2>📊 Current Store State</h2>
        <p>Real-time view of the store data:</p>

        <div class="state-grid">
          <!-- Auth Store State -->
          <div class="state-card store-card-hover" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
            <h3>🔐 Auth Store</h3><template x-if="$auth.isLoggedIn">
            <p><strong>Status:</strong> ✅ Logged In</p>
            <p><strong>User:</strong> </p>
            <p><strong>Email:</strong> </p></template><template x-else>
            <p><strong>Status:</strong> ❌ Logged Out</p>
            <p><em>Please log in to access your account</em></p></template>
            <small style="color: #6b7280">💡 Hover to see store object</small>
            <div class="tech-details">
              <pre class="state-display">isLoggedIn: <span x-text="$store.auth.isLoggedIn" style="color: #10b981;"></span>
user: <span x-text="$store.auth.user ? $store.auth.user.name : 'null'" style="color: #10b981;"></span></pre>
            </div>
          </div>

          <!-- Cart Store State -->
          <div class="state-card store-card-hover" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
            <h3>🛒 Cart Store</h3><template x-if="$cart.items.length > 0">
            <p><strong>Items:</strong> </p>
            <p><strong>Total:</strong> $</p>
            <div class="cart-items-list">
              <strong>Cart Contents:</strong>
              <ul><template x-for="item,  in $cart.items">
                <li>${item.name} - $${item.price}</li></template>
              </ul>
            </div></template><template x-else>
            <p><strong>Items:</strong> 0</p>
            <p><strong>Total:</strong> $0.00</p>
            <p><em>Cart is empty</em></p></template>
            <small style="color: #6b7280">💡 Hover to see store object</small>
            <div class="tech-details">
              <pre class="state-display">items: <span x-text="'[' + $store.cart.items.length + ' items]'" style="color: #10b981;"></span>
total: $<span x-text="$store.cart.formattedTotal" style="color: #10b981;"></span></pre>
            </div>
          </div>

          <!-- Theme Store State -->
          <div class="state-card store-card-hover" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
            <h3>🎨 Theme Store</h3><template x-if="$theme.mode === 'light'">
            <p><strong>Mode:</strong> ☀️ Light</p></template><template x-else>
            <p><strong>Mode:</strong> 🌙 Dark</p></template>
            <p>
              <strong>Background:</strong>
              <span x-text="$store.theme.getCurrentColors().background"></span>
            </p>
            <p>
              <strong>Text Color:</strong>
              <span x-text="$store.theme.getCurrentColors().text"></span>
            </p>
            <small style="color: #6b7280">💡 Hover to see store object</small>
            <div class="tech-details">
              <pre class="state-display">mode: "<span x-text="$store.theme.mode" style="color: #10b981;"></span>"
background: "<span x-text="$store.theme.getCurrentColors().background" style="color: #10b981;"></span>"
text: "<span x-text="$store.theme.getCurrentColors().text" style="color: #10b981;"></span>"</pre>
            </div>
          </div>
        </div>
      </div>

      <!-- Documentation Section -->
      <div class="docs-section" :style="\`background: ${$store.theme.mode === 'light' ? 'white' : '#2d3748'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
        <h2>📚 How It Works</h2>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>1. Store Files</h3>
          <p>Stores are defined in <code>stores/</code> directory:</p>
          <ul>
            <li><code>stores/auth.js</code> - Authentication state</li>
            <li><code>stores/cart.js</code> - Shopping cart state</li>
            <li><code>stores/theme.js</code> - Theme preferences</li>
          </ul>
        </div>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>2. Components Reference Stores</h3>
          <p>
            Components declare which stores they use in their fence section:
          </p>
          <pre class="code-display">
---
import store from './stores/auth.js'
---

<template x-if="$auth.isLoggedIn">
  Welcome, !</template></pre>
        </div>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>3. Automatic Reactivity</h3>
          <p>When any component modifies a store:</p>
          <ol>
            <li>Alpine.js detects the change</li>
            <li>All components using that store automatically update</li>
            <li>No manual event handling needed!</li>
          </ol>
        </div>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'};\`">
          <h3>4. Global Scope</h3>
          <p>Access stores anywhere with <code>$store.storeName</code>:</p>
          <ul>
            <li><code>@click="$store.auth.login()"</code> - Call methods</li>
            <li><code></code> - Read properties</li>
            <li>
              <code>{if $cart.items.length > 0}</code> - Use in conditionals
            </li>
            <li><code>{for item in $cart.items}</code> - Iterate arrays</li>
          </ul>
        </div>

        <div class="doc-card" :style="\`background: ${$store.theme.mode === 'light' ? '#f9fafb' : '#1a202c'}; border-color: ${$store.theme.mode === 'light' ? '#e5e7eb' : '#4a5568'}; border-left: 4px solid #10b981;\`">
          <h3 style="color: #10b981">5. Theme Reactivity</h3>
          <p>
            Notice how all elements on this page use
            <code>:style</code> bindings:
          </p>
          <pre class="code-display">
:style="&#96;background: &#36;&#123;&#36;store.theme.getCurrentColors().background&#125;;
         color: &#36;&#123;&#36;store.theme.getCurrentColors().text&#125;;&#96;"</pre>
          <p style="margin-top: 0.5rem">
            This makes the entire page reactive to theme changes!
          </p>
        </div>
      </div>
    </div>
  </body>
</html>`,
  'ThemeToggle': (props) => `<div class="theme-toggle">
  <span class="theme-label">Theme:</span>
  <div class="toggle-buttons">
    <button @click="$store.theme.setLight()" class="theme-btn" :class="{ active: $store.theme.mode === 'light' }">
      ☀️ Light
    </button>
    <button @click="$store.theme.setDark()" class="theme-btn" :class="{ active: $store.theme.mode === 'dark' }">
      🌙 Dark
    </button>
  </div>
  <button @click="$store.theme.toggle()" class="toggle-btn" title="Toggle theme">
    🔄
  </button>
</div>`,
  '../components/headerlogo.html': (props) => `<header class="header">
  <div class="header-container">
    <a href="/" class="brand">
      <svg width="36.206mm" height="7.781mm" version="1.1" viewBox="0 0 36.206 7.781" xmlns="http://www.w3.org/2000/svg">
        <g transform="translate(-264.95 11.104)">
          <path d="m273.59-6.453-2e-3 1.6107c-7e-3 0.22466-0.12785 0.35734-0.35253 0.36241-0.25549 0-0.36051-0.14631-0.36051-0.36123v-4.6353c0-0.092936 0.0232-0.16264 0.0696-0.20911 0.0464-0.052254 0.11906-0.092936 0.2178-0.12198 0.16852-0.052248 0.37175-0.090037 0.6099-0.11327 0.24404-0.023229 0.47924-0.034889 0.70576-0.034889 0.7377 0 1.2751 0.15683 1.612 0.4705 0.34263 0.30786 0.51403 0.72609 0.51403 1.2547 0 0.54021-0.17425 0.97296-0.52275 1.2982-0.34852 0.31948-0.88589 0.47922-1.612 0.47922zm0.83646-0.59249c0.46468 0 0.82484-0.095841 1.0804-0.28753 0.25564-0.19168 0.38343-0.48792 0.38343-0.88872 0-0.39499-0.12489-0.68252-0.37466-0.86259-0.24395-0.18588-0.59546-0.27882-1.0543-0.27882-0.1568 0-0.31363 0.00874-0.47042 0.026165-0.15106 0.011654-0.2876 0.02902-0.40959 0.052251v2.2393z">
          <path d="m290.02-4.9895c0.13651 0 0.23267-0.00934 0.28845-0.027942 0.062-0.024852 0.12724-0.037205 0.19548-0.037205 0.068 0 0.13031 0.027942 0.18614 0.083763 0.0561 0.049626 0.0837 0.12719 0.0837 0.23267 0 0.099268-0.0499 0.17062-0.1489 0.21405-0.21711 0.099268-0.40323 0.14891-0.55839 0.14891-0.15512 0-0.29778-0.00934-0.42812-0.027942-0.12403-0.012426-0.24506-0.052715-0.36289-0.12099-0.27922-0.16132-0.41883-0.44672-0.41883-0.8562v-2.3639h-0.53971c-0.0931 0-0.24964-0.10621-0.25249-0.29294-2e-3 -0.17599 0.15385-0.26609 0.23168-0.26866l0.55806-0.00462 0.0103-0.64389c5e-3 -0.12544 0.16669-0.25549 0.33632-0.25619 0.18493 7.719e-4 0.34881 0.12909 0.3458 0.26617v0.60493h0.86549c0.0871 0 0.15818 0.027942 0.21402 0.083754 0.056 0.055807 0.0837 0.12719 0.0837 0.21405 0 0.086862-0.0279 0.15821-0.0837 0.21405-0.0562 0.055807-0.12713 0.083761-0.21402 0.083761h-0.86549v2.308c0 0.19234 0.0588 0.32263 0.17683 0.39087 0.062 0.037205 0.16125 0.055884 0.29781 0.055884z">
          <g transform="matrix(.77185 0 0 .77185 312.99 -65.312)" fill-rule="evenodd" stroke-width=".26458px">
            <path d="m-54.57 74.523c-1.7919 1.907-5.3906 2.3551-4.3001 4.9881 0.36199 0.74862-0.27419 1.0155-0.50107 0.61951 0 0-3.4086-2.8786-1.4504-4.6156 0.98084-1.0161 6.9801-5.3003 6.3041-1.3029z" fill="#1c7fc7">
            <path d="m-57.629 73.238s-1.5075 0.83503-2.7039 1.8531c0 0 3.0859-1.3869 5.7631-0.56799 0 0 0.39145-1.6984-3.0593-1.2851z" fill="#004a77" fill-opacity=".87434">
            <path d="m-62.252 72.183s0.38056-1.1878 2.1798-1.9508l0.96418 0.28985s6.7226-1.3794 4.5374 4.0016c0 0 0.5771-2.071-4.5695-1.033l-0.70199 0.43123c-1.906-0.26683-2.4099-1.7389-2.4099-1.7389" fill="#22a6ed">
          </g>
          <path d="m277.62-9.6188q0-0.14572 0.10242-0.2368 0.10234-0.10019 0.25131-0.10019 0.14885 0 0.24197 0.10019 0.10234 0.091077 0.10234 0.2368v4.8453q0 0.13662-0.10234 0.2368-0.0931 0.10019-0.24197 0.10019-0.14898 0-0.25131-0.10019-0.10242-0.10018-0.10242-0.2368z">
          <path d="m280.03-6.1675q0.0651 0.51186 0.39083 0.81897 0.33506 0.29781 0.91202 0.29781 0.58636 0 0.99581-0.18613 0.13033-0.065145 0.23266-0.065145 0.10242-0.00934 0.18612 0.074482 0.0931 0.083761 0.0931 0.18613 0 0.20474-0.16752 0.2885-0.15823 0.083758-0.28845 0.14891-0.13034 0.065144-0.27927 0.11168-0.34429 0.093063-0.80033 0.093063-0.93997 0-1.4611-0.52116-0.51184-0.53048-0.51184-1.4984 0-0.83759 0.43737-1.396 0.49331-0.62354 1.3866-0.62354 0.85624 0 1.3495 0.577 0.46533 0.53978 0.46533 1.3494 0 0.1396-0.10234 0.24197-0.0931 0.10237-0.24196 0.10237zm1.126-1.6845q-0.67936 0-0.9865 0.60492-0.11163 0.21405-0.13952 0.51186h2.2521q-0.0277-0.53978-0.4002-0.85621-0.2978-0.26058-0.72589-0.26058z">
          <path d="m285.85-7.7962q-0.67009 0-1.2564 0.73522v2.2801q0 0.1396-0.10244 0.24197-0.10232 0.10237-0.24199 0.10237-0.13951 0-0.24199-0.10237-0.10232-0.10237-0.10232-0.24197v-3.2666q0-0.14891 0.093-0.25128 0.10246-0.10237 0.2513-0.10237 0.13964 0 0.24199 0.10237 0.10244 0.093062 0.10244 0.25128v0.31642q0.30708-0.31642 0.53042-0.45602 0.42814-0.25128 0.81904-0.25128 0.40014 0 0.65139 0.12099 0.25126 0.11167 0.42812 0.32573 0.35368 0.41879 0.35368 1.0423v2.1684q0 0.1396-0.10242 0.24197-0.10234 0.10237-0.24195 0.10237-0.13958 0-0.24199-0.10237-0.10234-0.10237-0.10234-0.24197v-2.094q0-0.42811-0.19544-0.67007-0.19549-0.25127-0.64217-0.25127z">
          <path d="m291.49-8.0475q0-0.15821 0.10242-0.25128 0.10234-0.10237 0.25127-0.10237 0.14887 0 0.24197 0.10237 0.093 0.093062 0.093 0.25128v3.2666q0 0.1396-0.093 0.24197-0.0931 0.10237-0.24197 0.10237-0.14894 0-0.25127-0.10237-0.10242-0.10237-0.10242-0.24197zm0.76312-1.4983q0 0.16751-0.11161 0.27919-0.11174 0.11168-0.27921 0.11168h-0.0465q-0.16751 0-0.27921-0.11168-0.10234-0.11168-0.10234-0.27919v-0.027942q0-0.16752 0.11165-0.2792 0.11167-0.11167 0.26989-0.11167h0.0465q0.16747 0 0.27921 0.11167 0.11161 0.11168 0.11161 0.2792z">
          <path d="m295.26-5.0508q0.42813 0 0.80971-0.20474 0.13025-0.074483 0.24196-0.074483 0.11166-0.00934 0.19538 0.093062 0.0839 0.10237 0.0839 0.24197 0 0.1303-0.17686 0.24198-0.53977 0.35365-1.2098 0.35365-0.84694 0-1.4332-0.53047-0.61426-0.5677-0.60495-1.489 0-0.92135 0.60495-1.489 0.57697-0.53978 1.4332-0.53047 0.45601 0 0.74448 0.12099 0.29781 0.11167 0.46533 0.23266 0.17686 0.11168 0.17686 0.25128 0 0.1396-0.0839 0.23267-0.0837 0.093062-0.16746 0.093062-0.13033 0-0.26988-0.074483-0.39088-0.21405-0.80971-0.20474-0.67006 0-1.0423 0.38156-0.363 0.37226-0.363 0.98649 0 0.61423 0.363 0.99579 0.37224 0.37226 1.0423 0.37226z">
          <path d="m299.21-8.4383q0.88411 0 1.4239 0.55839 0.53971 0.5677 0.52112 1.4611 0 0.90273-0.52112 1.4611-0.53979 0.55839-1.4239 0.55839-0.89347 0-1.4146-0.55839-0.53979-0.55839-0.53979-1.4611 0-0.91204 0.53979-1.4611 0.52111-0.55839 1.4146-0.55839zm-0.85627 3.0991q0.18617 0.15821 0.40954 0.23267 0.23266 0.074482 0.44673 0.074482 0.21402 0 0.43739-0.074482 0.22337-0.074483 0.40943-0.24197 0.4095-0.37226 0.4095-1.0796 0-0.68868-0.4095-1.0702-0.34429-0.31642-0.84682-0.30712-0.82832 0-1.1447 0.74453-0.11171 0.26988-0.11171 0.64215 0 0.37226 0.11171 0.64215 0.11164 0.26989 0.28846 0.43741z">
        </g>
      </svg>
    </a>

    <nav class="nav">
      <a href="/" class="nav-item">Home</a>
      <a href="/products" class="nav-item">Products</a>
      <a href="/login" class="nav-item">Login</a>
      <a href="/register" class="nav-item">Register</a>
    </nav>
  </div>
</header>`,
  '../components/header.html': (props) => `<header class="header">
  <div class="header-container">
    <div>
      <a href="/" class="brand">
        <svg width="36.206mm" height="7.781mm" version="1.1" viewBox="0 0 36.206 7.781" xmlns="http://www.w3.org/2000/svg">
          <g transform="translate(-264.95 11.104)">
            <path d="m273.59-6.453-2e-3 1.6107c-7e-3 0.22466-0.12785 0.35734-0.35253 0.36241-0.25549 0-0.36051-0.14631-0.36051-0.36123v-4.6353c0-0.092936 0.0232-0.16264 0.0696-0.20911 0.0464-0.052254 0.11906-0.092936 0.2178-0.12198 0.16852-0.052248 0.37175-0.090037 0.6099-0.11327 0.24404-0.023229 0.47924-0.034889 0.70576-0.034889 0.7377 0 1.2751 0.15683 1.612 0.4705 0.34263 0.30786 0.51403 0.72609 0.51403 1.2547 0 0.54021-0.17425 0.97296-0.52275 1.2982-0.34852 0.31948-0.88589 0.47922-1.612 0.47922zm0.83646-0.59249c0.46468 0 0.82484-0.095841 1.0804-0.28753 0.25564-0.19168 0.38343-0.48792 0.38343-0.88872 0-0.39499-0.12489-0.68252-0.37466-0.86259-0.24395-0.18588-0.59546-0.27882-1.0543-0.27882-0.1568 0-0.31363 0.00874-0.47042 0.026165-0.15106 0.011654-0.2876 0.02902-0.40959 0.052251v2.2393z">
            <path d="m290.02-4.9895c0.13651 0 0.23267-0.00934 0.28845-0.027942 0.062-0.024852 0.12724-0.037205 0.19548-0.037205 0.068 0 0.13031 0.027942 0.18614 0.083763 0.0561 0.049626 0.0837 0.12719 0.0837 0.23267 0 0.099268-0.0499 0.17062-0.1489 0.21405-0.21711 0.099268-0.40323 0.14891-0.55839 0.14891-0.15512 0-0.29778-0.00934-0.42812-0.027942-0.12403-0.012426-0.24506-0.052715-0.36289-0.12099-0.27922-0.16132-0.41883-0.44672-0.41883-0.8562v-2.3639h-0.53971c-0.0931 0-0.24964-0.10621-0.25249-0.29294-2e-3 -0.17599 0.15385-0.26609 0.23168-0.26866l0.55806-0.00462 0.0103-0.64389c5e-3 -0.12544 0.16669-0.25549 0.33632-0.25619 0.18493 7.719e-4 0.34881 0.12909 0.3458 0.26617v0.60493h0.86549c0.0871 0 0.15818 0.027942 0.21402 0.083754 0.056 0.055807 0.0837 0.12719 0.0837 0.21405 0 0.086862-0.0279 0.15821-0.0837 0.21405-0.0562 0.055807-0.12713 0.083761-0.21402 0.083761h-0.86549v2.308c0 0.19234 0.0588 0.32263 0.17683 0.39087 0.062 0.037205 0.16125 0.055884 0.29781 0.055884z">
            <g transform="matrix(.77185 0 0 .77185 312.99 -65.312)" fill-rule="evenodd" stroke-width=".26458px">
              <path d="m-54.57 74.523c-1.7919 1.907-5.3906 2.3551-4.3001 4.9881 0.36199 0.74862-0.27419 1.0155-0.50107 0.61951 0 0-3.4086-2.8786-1.4504-4.6156 0.98084-1.0161 6.9801-5.3003 6.3041-1.3029z" fill="#1c7fc7">
              <path d="m-57.629 73.238s-1.5075 0.83503-2.7039 1.8531c0 0 3.0859-1.3869 5.7631-0.56799 0 0 0.39145-1.6984-3.0593-1.2851z" fill="#004a77" fill-opacity=".87434">
              <path d="m-62.252 72.183s0.38056-1.1878 2.1798-1.9508l0.96418 0.28985s6.7226-1.3794 4.5374 4.0016c0 0 0.5771-2.071-4.5695-1.033l-0.70199 0.43123c-1.906-0.26683-2.4099-1.7389-2.4099-1.7389" fill="#22a6ed">
            </g>
            <path d="m277.62-9.6188q0-0.14572 0.10242-0.2368 0.10234-0.10019 0.25131-0.10019 0.14885 0 0.24197 0.10019 0.10234 0.091077 0.10234 0.2368v4.8453q0 0.13662-0.10234 0.2368-0.0931 0.10019-0.24197 0.10019-0.14898 0-0.25131-0.10019-0.10242-0.10018-0.10242-0.2368z">
            <path d="m280.03-6.1675q0.0651 0.51186 0.39083 0.81897 0.33506 0.29781 0.91202 0.29781 0.58636 0 0.99581-0.18613 0.13033-0.065145 0.23266-0.065145 0.10242-0.00934 0.18612 0.074482 0.0931 0.083761 0.0931 0.18613 0 0.20474-0.16752 0.2885-0.15823 0.083758-0.28845 0.14891-0.13034 0.065144-0.27927 0.11168-0.34429 0.093063-0.80033 0.093063-0.93997 0-1.4611-0.52116-0.51184-0.53048-0.51184-1.4984 0-0.83759 0.43737-1.396 0.49331-0.62354 1.3866-0.62354 0.85624 0 1.3495 0.577 0.46533 0.53978 0.46533 1.3494 0 0.1396-0.10234 0.24197-0.0931 0.10237-0.24196 0.10237zm1.126-1.6845q-0.67936 0-0.9865 0.60492-0.11163 0.21405-0.13952 0.51186h2.2521q-0.0277-0.53978-0.4002-0.85621-0.2978-0.26058-0.72589-0.26058z">
            <path d="m285.85-7.7962q-0.67009 0-1.2564 0.73522v2.2801q0 0.1396-0.10244 0.24197-0.10232 0.10237-0.24199 0.10237-0.13951 0-0.24199-0.10237-0.10232-0.10237-0.10232-0.24197v-3.2666q0-0.14891 0.093-0.25128 0.10246-0.10237 0.2513-0.10237 0.13964 0 0.24199 0.10237 0.10244 0.093062 0.10244 0.25128v0.31642q0.30708-0.31642 0.53042-0.45602 0.42814-0.25128 0.81904-0.25128 0.40014 0 0.65139 0.12099 0.25126 0.11167 0.42812 0.32573 0.35368 0.41879 0.35368 1.0423v2.1684q0 0.1396-0.10242 0.24197-0.10234 0.10237-0.24195 0.10237-0.13958 0-0.24199-0.10237-0.10234-0.10237-0.10234-0.24197v-2.094q0-0.42811-0.19544-0.67007-0.19549-0.25127-0.64217-0.25127z">
            <path d="m291.49-8.0475q0-0.15821 0.10242-0.25128 0.10234-0.10237 0.25127-0.10237 0.14887 0 0.24197 0.10237 0.093 0.093062 0.093 0.25128v3.2666q0 0.1396-0.093 0.24197-0.0931 0.10237-0.24197 0.10237-0.14894 0-0.25127-0.10237-0.10242-0.10237-0.10242-0.24197zm0.76312-1.4983q0 0.16751-0.11161 0.27919-0.11174 0.11168-0.27921 0.11168h-0.0465q-0.16751 0-0.27921-0.11168-0.10234-0.11168-0.10234-0.27919v-0.027942q0-0.16752 0.11165-0.2792 0.11167-0.11167 0.26989-0.11167h0.0465q0.16747 0 0.27921 0.11167 0.11161 0.11168 0.11161 0.2792z">
            <path d="m295.26-5.0508q0.42813 0 0.80971-0.20474 0.13025-0.074483 0.24196-0.074483 0.11166-0.00934 0.19538 0.093062 0.0839 0.10237 0.0839 0.24197 0 0.1303-0.17686 0.24198-0.53977 0.35365-1.2098 0.35365-0.84694 0-1.4332-0.53047-0.61426-0.5677-0.60495-1.489 0-0.92135 0.60495-1.489 0.57697-0.53978 1.4332-0.53047 0.45601 0 0.74448 0.12099 0.29781 0.11167 0.46533 0.23266 0.17686 0.11168 0.17686 0.25128 0 0.1396-0.0839 0.23267-0.0837 0.093062-0.16746 0.093062-0.13033 0-0.26988-0.074483-0.39088-0.21405-0.80971-0.20474-0.67006 0-1.0423 0.38156-0.363 0.37226-0.363 0.98649 0 0.61423 0.363 0.99579 0.37224 0.37226 1.0423 0.37226z">
            <path d="m299.21-8.4383q0.88411 0 1.4239 0.55839 0.53971 0.5677 0.52112 1.4611 0 0.90273-0.52112 1.4611-0.53979 0.55839-1.4239 0.55839-0.89347 0-1.4146-0.55839-0.53979-0.55839-0.53979-1.4611 0-0.91204 0.53979-1.4611 0.52111-0.55839 1.4146-0.55839zm-0.85627 3.0991q0.18617 0.15821 0.40954 0.23267 0.23266 0.074482 0.44673 0.074482 0.21402 0 0.43739-0.074482 0.22337-0.074483 0.40943-0.24197 0.4095-0.37226 0.4095-1.0796 0-0.68868-0.4095-1.0702-0.34429-0.31642-0.84682-0.30712-0.82832 0-1.1447 0.74453-0.11171 0.26988-0.11171 0.64215 0 0.37226 0.11171 0.64215 0.11164 0.26989 0.28846 0.43741z">
          </g>
        </svg>
      </a>
    </div>
    
    <nav class="nav"><template x-for="item,  in navItems">
        <a href="${item.url}" class="nav-item">${item.label}</a></template>
    </nav><template x-if="isLoggedIn">
      <div class="user-info">
        <span>Welcome, ${props.user.name}</span><template x-if="user.role === "admin"">
          <span class="user-badge admin-badge">Admin</span></template><template x-else>
          <span class="user-badge">${props.user.role}</span></template>
      </div></template>
  </div>
</header>`,
  '../global/nav.html': (props) => `<!-- ============================================ --><!--                 Navigation                   --><!-- ============================================ --><header id="cs-navigation" :class="{'cs-active': mobileMenuOpen}">
    <div class="cs-container">
        <div class="cs-top">
            <!--Navigation List-->
            <nav class="cs-nav" aria-label="main navigation">
                <!--Mobile Nav Toggle-->
                <button class="cs-toggle" @click="toggleMobileMenu()" :aria-expanded="mobileMenuOpen" aria-label="mobile menu toggle" aria-controls="cs-ul-wrapper" aria-haspopup="menu">
                    <span class="cs-box" aria-hidden="true">
                        <span class="cs-line cs-line1"></span>
                        <span class="cs-line cs-line2"></span>
                        <span class="cs-line cs-line3"></span>
                    </span>
                </button>
                <!-- We need a wrapper div so we can set a fixed height on the cs-ul in case the nav list gets too long from too many dropdowns being opened and needs to have an overflow scroll. This wrapper acts as the background so it can go the full height of the screen and not cut off any overflowing nav items while the cs-ul stops short of the bottom of the screen, which keeps all nav items in view no matter how mnay there are-->
                <div class="cs-ul-wrapper" id="cs-ul-wrapper">
                    <ul class="cs-ul">
                        <li class="cs-li">
                            <a href="" class="cs-li-link cs-active" aria-current="page">
                                Home
                            </a>
                        </li>
                        <li class="cs-li">
                            <a href="" class="cs-li-link">
                                About
                            </a>
                        </li>
                        <!--Copy and paste this cs-dropdown list item and replace any .cs-li with this cs-dropdown group to make a new dropdown and it will work-->
                        <!-- Note: When copying this dropdown, update the id attribute with an appropriate name (e.g., "services-dropdown-toggle", "products-dropdown-toggle") and ensure the aria-labelledby attribute on the ul matches the button's id -->
                        <li class="cs-li cs-dropdown">
                            <button class="cs-li-link cs-dropdown-toggle" aria-expanded="false" aria-haspopup="menu" id="services-dropdown-toggle">
                                Services
                                <img class="cs-drop-icon" src="https://csimages2.nyc3.digitaloceanspaces.com/Icons/chevron-dropdown.svg" alt="" width="15" height="15" decoding="async" aria-hidden="true"></img>
                            </button>
                            <ul class="cs-drop-ul" aria-labelledby="services-dropdown-toggle">
                                <li class="cs-drop-li">
                                    <a href="" class="cs-li-link cs-drop-link">Service 1</a>
                                </li>
                                <li class="cs-drop-li">
                                    <a href="" class="cs-li-link cs-drop-link">Service 2</a>
                                </li>
                            </ul>
                        </li>
                    </ul>
                    <ul class="cs-ul">
                        <li class="cs-li cs-dropdown">
                            <button class="cs-li-link cs-dropdown-toggle" aria-expanded="false" aria-haspopup="menu" id="services-dropdown-toggle">
                                Artists
                                <img class="cs-drop-icon" src="https://csimages2.nyc3.digitaloceanspaces.com/Icons/chevron-dropdown.svg" alt="" width="15" height="15" decoding="async" aria-hidden="true"></img>
                            </button>
                            <ul class="cs-drop-ul" aria-labelledby="services-dropdown-toggle">
                                <li class="cs-drop-li">
                                    <a href="" class="cs-li-link cs-drop-link">Artist 1</a>
                                </li>
                                <li class="cs-drop-li">
                                    <a href="" class="cs-li-link cs-drop-link">Artist 2</a>
                                </li>
                            </ul>
                        </li>
                        <li class="cs-li">
                            <a href="" class="cs-li-link">FAQ</a>
                        </li>
                        <li class="cs-li">
                            <a href="" class="cs-li-link">Contact</a>
                        </li>
                    </ul>
                </div>
                <!--Dark Mode toggle, uncomment button code if you want to enable a dark mode toggle-->
                <button id="dark-mode-toggle" aria-label="dark mode toggl" aria-pressed="false">
                    <svg class="cs-moon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 480 480" style="enable-background:new 0 0 480 480" xml:space="preserve" aria-hidden="true">
                        <path d="M459.782 347.328c-4.288-5.28-11.488-7.232-17.824-4.96-17.76 6.368-37.024 9.632-57.312 9.632-97.056 0-176-78.976-176-176 0-58.4 28.832-112.768 77.12-145.472 5.472-3.712 8.096-10.4 6.624-16.832S285.638 2.4 279.078 1.44C271.59.352 264.134 0 256.646 0c-132.352 0-240 107.648-240 240s107.648 240 240 240c84 0 160.416-42.688 204.352-114.176 3.552-5.792 3.04-13.184-1.216-18.496z">
                    </svg>
                    <img class="cs-sun" aria-hidden="true" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Icons%2Fsun.svg" decoding="async" alt="" width="15" height="15"></img>
                </button>
            </nav>

        </div>
        <div class="cs-bottom">
            <!--Nav Logo-->
            <a href="" class="cs-logo" aria-label="Return to [Company Name]home page">
                <img src="https://csimages2.nyc3.digitaloceanspaces.com/Icons/artistitch-logo.svg" alt="" width="210" height="29" aria-hidden="true" decoding="async"></img>
            </a>
        </div>
    </div>
</header>`,
  '../content/nav-test.html': (props) => `<html lang="en">
<head>
	<meta charset="UTF-8"></meta>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"></meta>
	<title>${props.title}</title>

	<!-- Alpine.js - Global reactive framework -->
	<script src="//unpkg.com/alpinejs" defer=""></script>

	<!-- App Styles -->
	<link rel="stylesheet" href="/styles/style.css"></link>
</head>
<body>

<!-- Import and render Nav component -->

<!-- Test content -->
<div style="padding: 2rem; max-width: 800px; margin: 0 auto;">
	<h1>Navigation Component Test</h1>
	<p>If you see the navigation above, the Nav component is working correctly!</p>

	<h2>Test Checklist:</h2>
	<ul>
		<li>✓ Navigation bar appears at top</li>
		<li>✓ Logo is visible</li>
		<li>✓ Menu items (Home, About, Services, Artists, FAQ, Contact) are visible</li>
		<li>✓ Dark mode toggle button appears</li>
		<li>✓ Dropdown menus work (Services, Artists)</li>
		<li>✓ Mobile hamburger menu works on small screens</li>
	</ul>

	<h2>How to Test:</h2>
	<ol>
		<li>Resize your browser window to test responsive behavior</li>
		<li>Click on "Services" or "Artists" to test dropdown menus</li>
		<li>Click the dark mode toggle (moon/sun icon)</li>
		<li>On mobile width, click the hamburger menu (three lines)</li>
	</ol>
</div>

</body>
</html>`,
  'Pages': (props) => `<!-- Pages Layout - Renders dynamic components from content JSON --><!-- NOTE: <head> provided by wrapper layout (includes Alpine.js + runtime-components.js) --><!-- NOTE: No fence section - uses variables from parent scope (html.html wrapper) --><template x-for="component,  in content.components"></template>`,
  'Services2437': (props) => `<!-- ============================================ --><!--                  Services                    --><!-- ============================================ --><section id="services-2437">
	<div class="cs-container">
		<div class="cs-content">
			<span class="cs-topper">Main Services</span>
			<h2 class="cs-title">Explore More Ways to Glow and Express Yourself</h2>
		</div>
		<ul class="cs-card-group">
			<li class="cs-item">
				<a href="" class="cs-link">
					<h3 class="cs-h3">Custom UV Body Painting</h3>
					<picture class="cs-picture">
						<!--Mobile Image-->
						<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art5.jpg">
						<!--Tablet and above Image-->
						<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art5.jpg">
						<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art5.jpg" alt="tattoo" width="373" height="420">
					</picture>
					<p class="cs-item-text">
                        High-definition, blacklight-reactive designs for events, portraits, or personal expression.
                    </p>
                    <div class="cs-border cs-border1" aria-hidden="true"></div>
                    <div class="cs-border cs-border2" aria-hidden="true"></div>
				</a>
			</li>
			<li class="cs-item">
				<a href="" class="cs-link">
					<h3 class="cs-h3">Spiritual & Symbolic Art</h3>
					<picture class="cs-picture">
						<!--Mobile Image-->
						<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art6.jpg">
						<!--Tablet and above Image-->
						<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art6.jpg">
						<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art6.jpg" alt="room" width="373" height="420">
					</picture>
					<p class="cs-item-text">
                        Sacred geometry, mythic creatures, abstract motifs, and personal symbols
                    </p>
                    <div class="cs-border cs-border1" aria-hidden="true"></div>
                    <div class="cs-border cs-border2" aria-hidden="true"></div>
				</a>
			</li>
			<li class="cs-item">
				<a href="" class="cs-link">
					<h3 class="cs-h3">Workshops & Education</h3>
					<picture class="cs-picture">
						<!--Mobile Image-->
						<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art7.jpg">
						<!--Tablet and above Image-->
						<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art7.jpg">
						<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art7.jpg" alt="tattoo" width="373" height="420">
					</picture>
					<p class="cs-item-text">
                        Beginner-friendly body art training and creative process sharing
                    </p>
                    <div class="cs-border cs-border1" aria-hidden="true"></div>
                    <div class="cs-border cs-border2" aria-hidden="true"></div>
				</a>
			</li>
            <li class="cs-item">
				<a href="" class="cs-link">
					<h3 class="cs-h3">Workshops & Education</h3>
					<picture class="cs-picture">
						<!--Mobile Image-->
						<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art8.jpg">
						<!--Tablet and above Image-->
						<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art8.jpg">
						<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/People/art8.jpg" alt="tattoo" width="373" height="420">
					</picture>
					<p class="cs-item-text">
                        Beginner-friendly body art training and creative process sharing
                    </p>
                    <div class="cs-border cs-border1" aria-hidden="true"></div>
                    <div class="cs-border cs-border2" aria-hidden="true"></div>
				</a>
			</li>
		</ul>
        <a href="" class="cs-button-solid">Explore All Services</a>
	</div>
    <picture class="cs-background">
		<source media="(max-width: 600px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/Graphics/paint.png">
		<source media="(min-width: 601px)" srcset="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/Graphics/paint.png">
		<img loading="lazy" decoding="async" src="https://csimg.nyc3.cdn.digitaloceanspaces.com/Images/Graphics/paint.png" alt="splatter" width="1920" height="1168" aria-hidden="true">
	</picture>
</section>`
};