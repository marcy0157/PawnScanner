document.addEventListener('DOMContentLoaded', () => {
    console.log("JavaScript caricato correttamente");

    const emailForm = document.getElementById('emailForm');
    const emailInput = document.getElementById('emailInput');
    const resultsDiv = document.getElementById('results');
    const heroSection = document.querySelector('.hero-section');
    const resultsSection = document.querySelector('.results-section');

    emailForm.addEventListener('submit', async (event) => {
        event.preventDefault(); // Previene il ricaricamento della pagina

        const emailValue = emailInput.value.trim();

        // Controllo se l'email è vuota
        if (!emailValue) {
            alert("Inserisci un'email valida!");
            return;
        }

        console.log("Email inserita:", emailValue);

        // Pulisce i risultati precedenti
        resultsDiv.innerHTML = `
            <p class="text-center animate__animated animate__fadeIn">Stiamo cercando nei database...</p>
        `;

        const breachImages = {
            'Facebook': '/media/img/facebook.png',
            'LinkedIn': '/media/img/linkedin.png',
            'Twitter': '/media/img/twitter.png',
            'Adobe': '/media/img/adobe.png',
            'VK': '/media/img/vk.png',
            'Tumblr': '/media/img/tumblr.png',
            'Badoo': '/media/img/badoo.png',
            'Last.fm': '/media/img/lastfm.png',
            'Zynga': '/media/img/zynga.png',
            'Canva': '/media/img/canva.png',
            '500px': '/media/img/500px.png',
            'Disqus': '/media/img/disqus.png',
            'LiveJournal': '/media/img/livejournal.png',
            'MySpace': '/media/img/myspace.png',
            'Patreon': '/media/img/patreon.png',
            'Wattpad': '/media/img/wattpad.png',
            'Instagram': '/media/img/instagram.png',
            'Dropbox': '/media/img/dropbox.png',
            'Yahoo': '/media/img/yahoo.png',
            'Apple': '/media/img/apple.png',
            'Amazon': '/media/img/amazon.png',
            'Netflix': '/media/img/netflix.png',
            'Spotify': '/media/img/spotify.png',
            'Google': '/media/img/google.png',
            'PayPal': '/media/img/paypal.png',
            'eBay': '/media/img/ebay.png',
            'Uber': '/media/img/uber.png',
            'Airbnb': '/media/img/airbnb.png',
            'TikTok': '/media/img/tiktok.png',
            'Pinterest': '/media/img/pinterest.png',
            'Snapchat': '/media/img/snapchat.png',
            'Reddit': '/media/img/reddit.png',
            'Twitch': '/media/img/twitch.png',
            'GitHub': '/media/img/github.png',
            'Steam': '/media/img/steam.png',
            'Epic Games': '/media/img/epicgames.png',
            'HBO': '/media/img/hbo.png',
            'Slack': '/media/img/slack.png',
            'Microsoft': '/media/img/microsoft.png',
            'Nintendo': '/media/img/nintendo.png',
            'Tinder': '/media/img/tinder.png',
            'Vodafone': '/media/img/vodafone.png',
            'YouTube': '/media/img/youtube.png',
            'Xbox': '/media/img/xbox.png',
            'PlayStation': '/media/img/playstation.png',
            'WhatsApp': '/media/img/whatsapp.png',
            'Telegram': '/media/img/telegram.png',
            'Discord': '/media/img/discord.png',
            'Generic': '/media/img/generic.png' // Immagine generica per breach non specifici
        };


        try {
            // Chiamata API
            const response = await fetch('/check-email', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer YOUR_SECRET_TOKEN'
                },
                body: JSON.stringify({ email: emailValue })
            });

            if (response.ok) {
                const data = await response.json();

                if (data.breaches && data.breaches.length > 0) {
                    resultsDiv.innerHTML = `
                    <div class="custom-alert animate__animated animate__fadeIn">
                        <strong>Email trovata nei seguenti breach:</strong>
                        <ul class="list-group mt-3">
                            ${data.breaches.map(breach => `
                                <li class="list-group-item d-flex align-items-center">
                                    <img src="${breachImages[breach] || breachImages['Generic']}" alt="${breach}" class="me-3" style="width: 24px; height: 24px;">
                                    <span>${breach}</span>
                                </li>
                            `).join('')}
                        </ul>
                    </div>`;
                } else {
                    resultsDiv.innerHTML = `
                    <div class="alert alert-success animate__animated animate__fadeIn">
                        Nessun breach trovato per questa email.
                    </div>`;
                }
            } else {
                const errorData = await response.json();
                resultsDiv.innerHTML = `
                    <div class="alert alert-warning animate__animated animate__fadeIn">
                        ${errorData.message}
                    </div>`;
            }

            // Mostra i risultati
            resultsSection.style.display = 'block'; // Rimuove display: none
            resultsSection.classList.add('visible');
            heroSection.classList.add('reduced'); // Riduce l'altezza con un'animazione
        } catch (error) {
            console.error("Errore durante la chiamata API:", error);
            resultsDiv.innerHTML = `
                <div class="alert alert-danger animate__animated animate__fadeIn">
                    Si è verificato un errore: ${error.message}
                </div>`;
        }
    });
});
