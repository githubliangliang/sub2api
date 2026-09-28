# 实施证据校验

SHA-256按LF换行归一化后的UTF-8内容计算。代码文件内容另见code-manifest.tsv；本表不包含自身。

| 文件 | SHA-256 |
|---|---|
| alpha-status-green.txt | 42f17d4cb8102eb0bc2cec0c3cacdcd2f51f6e088a8f413823fba0a8a39630f4 |
| alpha-status-red.txt | a6958c6deb6cf034ce74dbea0cdf22140010d8d78abd133c3ffb97fd245e7fb4 |
| backend-green.txt | 18124d320ba86830fcafec5c4abb18619df3be4782df361aa93e460e5fb7c4b0 |
| backend-race.txt | 8b8ea8d295e9a0ab44b3ed4fbc96f3534d26260fd81786cecc03eeaac310ccb0 |
| backend-red-valid.txt | 5d10c00e273b6b6b5d60346b16e9149a50075ed10b2629752218b6afdd3d4813 |
| backend-report.md | 76399cae338744a4aa0547e2ddd71559ad9a8f360e3e8d95210fee88fd04a26f |
| baseline-build.txt | 01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b |
| baseline-list.txt | 2e4f783123d7aad99c09296ee3c238122ee9217e9d85e8aa8cf0bf7c5f95d48a |
| baseline-unit.txt | d93c8f68a2b0aff09d0d85e6b69e1600667cf35d08f587a3afa17980783d15ae |
| code-manifest.tsv | c8fdd82593459d212cac7db985d8f536d758d77c37b6fa1a7f7dd2be05e76e6a |
| completion-audit.json | f8c8d3832d7aba5f99a1ce3e86f5dbccf0c1c6ebd66fce6617e28332a95f3aef |
| final-backend-build.txt | 01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b |
| final-backend-unit.txt | 3b5aff32624295ab755b663327cba829461b204446e95c519cf7bf2ac28e09bc |
| final-frontend-build.txt | 694014d1d6cfdebe4ea443deadd2ebf433180e5d0551c9ab701c76ab67aab28e |
| final-frontend-unit.txt | 02b5e61a444fda729029340bfa63afd56b22b8190224392a321cad55d8019a18 |
| final-sqlite.txt | e8567923b103aa3e02fe8d296fd0e7d06d74185a25207dbab1b238584f6f9c42 |
| frontend-F09-green.txt | 9b284153dca912c082e26523227d9deeb5434b3099c054f07af626e1561cdf58 |
| frontend-F09-red.txt | 082cf1c704d1beb7f652cea962cdcade758222b9306a1d24b1f85875526ead0d |
| frontend-F10-green.txt | 3685c3207e64b893a6c67578b3c517bb1a40f5695056e8b5fba8ddc81ff84ea1 |
| frontend-F10-red.txt | 3dd7311715d0fbb8744b27b296da43b3bc3a6d73becd8ffb391657629c1069ed |
| frontend-F11-green.txt | 2497351246855ab494a2262bc7e14e2e5037af5f071fc809be075a9e946bf610 |
| frontend-F11-red.txt | b7601b7443616d0c2c688a44f6428af8b843ae448b9872a23f53fa8cbd47b61e |
| frontend-F12-draft-green-final.txt | 73f39b331a24513005f9a8736ff02a6ee46a4467bfca52e721d2f214bd0566ab |
| frontend-F12-draft-lint.txt | a57676ee64438c0e0d0b65228b829d156702cafdd7e2dfd578c9977a2c444b1b |
| frontend-F12-draft-red.txt | 65519596ca68aa693a90a3df32ea2f5bddadc6bfcdf3d378b37e91526ef58383 |
| frontend-F12-draft-typecheck.txt | a7c6f04f0e36a2de8db9c98735c3d2a9e4b2912e1a95a38a22456aa05357b6bb |
| frontend-F12-edit-green-final.txt | 51e766ad24a9efc2dcebfa065f76815b5acf7f6cefacf5b720831547f4e9b7bd |
| frontend-F12-edit-green.txt | d2375729924f2f99ae3e0e75279790e8aaeccd96bf1ac017674afd8fa39d851d |
| frontend-F12-edit-lint.txt | a57676ee64438c0e0d0b65228b829d156702cafdd7e2dfd578c9977a2c444b1b |
| frontend-F12-edit-native-red.txt | 83ad0889901ccc28eb7334baa8a6ad483e304bb84ade634606faa169b454d3c5 |
| frontend-F12-edit-red.txt | 2ef5d8e67dab727f7d8c57abe60227684c37d99c4c508d300e396997d463b895 |
| frontend-F12-edit-typecheck.txt | a7c6f04f0e36a2de8db9c98735c3d2a9e4b2912e1a95a38a22456aa05357b6bb |
| frontend-F12-green.txt | ad81ea749a21faafe44dcb11d18c35661d97185912ec25ef6b95866b2fb894c5 |
| frontend-F12-red.txt | 48589865b0b250eb4a9257d3b92ae5e48399fb32ac4cbab2fdd49a4dabb46cfa |
| frontend-F13-green.txt | 99da3ac184c88e64c8c2d5461e4efd64999e27c51d8ac491db73783ceeca16e9 |
| frontend-F13-red.txt | ddd6cc30b5a7103891c42eeaa5840f9c5614e384df9312bbcc6b41be8b33c71d |
| frontend-F14-green.txt | a7c4cbdbb72622163aa7e3a9b98d334b824542adac03a74168e49db50d45ff86 |
| frontend-F14-red.txt | 6099c82457296b75d112cd7247b4ae1bac8e43a33d2f1b9c96196e890facba0e |
| frontend-final-specs.txt | 60df2bc3df38b4ded15acf1f5616fd2396537dc5bfbef0e4d15140ac1bcdec84 |
| frontend-final-tests.txt | 2e725cf97c35e8df05f4094168cdd5998d4b4e9b7925ef7235610b5ac8267d8f |
| frontend-lint.txt | a57676ee64438c0e0d0b65228b829d156702cafdd7e2dfd578c9977a2c444b1b |
| frontend-manifest.json | 972580f27bb98be6b3f264bdf518142d354a4f62bd6b22a51967115c7cf66fa1 |
| frontend-report.md | 1e1c1c296f3ea0065ec427a16fc4c5a15689a61954f7af53dae12d8338354b4f |
| frontend-typecheck.txt | a7c6f04f0e36a2de8db9c98735c3d2a9e4b2912e1a95a38a22456aa05357b6bb |
| README.md | fc2407a7b12b7aa7c18926ed7806d9c3f857fcbb9a6213929448947d70b7d08f |
| review.md | 5f10e878d25883e13274d21c938816a014303912d1fba49c322c78766581126f |
| source-coverage.tsv | 2f88c55fa5247b6c7490f2f6cf856d3a9cc38d83ddfd2540702fe61c79ffc6e8 |
