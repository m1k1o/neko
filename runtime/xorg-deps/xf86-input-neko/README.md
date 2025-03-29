# xf86-input-neko
[X.org](https://x.org/) [neko](http://m1k1o/neko) input driver

### how to use
xf86-input-neko assumes you have only one virtual touchscreen device available, see
`80-neko.conf`. If there are multiple in your system, please specify one config
section for each.
xf86-input-neko aims to make [neko](http://m1k1o/neko) easy to use and doesn't
offer special configuration options.

* `./configure --prefix=/usr`
* `make`
* `sudo make install`

Done.

To _uninstall_, again go inside the extracted directory, and do

    sudo make uninstall
