# Source Baseline（冻结）

冻结日期2026-11-18。实施不得改写本文件；真实起点另填verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `7a0b78fe511f17b0f7eb9d26fcf5e4ff083312c7` / `3882c3ce52fb24f6f8d89d782ffc7b48bb1e65e7` |
| VERSION / Go / SQLite / auth snapshot | `1.1.25 / 1.27.3 / 229 / 21` |
| private index tree | `3882c3ce52fb24f6f8d89d782ffc7b48bb1e65e7` |
| v0.2.21 commit | `10ba8f9b4dc2e6023d4a9879ab77b21841babd10` |
| v0.2.22 commit / tag object | `b0fa9f3e6eeab1968a53e1289b69daa4bb40ca2b` / `b11573c9d55d39b6780e1e9f6e193ba2e6ca2321` |
| v0.2.23 commit / tag object | `5af9afb588c4b014b3df9db7d4073af1f8ae01b6` / `ce19a640eab5c4b5ef1827c7c1fcfce49569411a` |
| upstream/main（观察，不作来源） | `2af3f1e15c98c893a89255a037f551046883b8c0` |
| 非merge / 候选 / 净diff | 128 / 78 PR + 4直接提交 / 399文件 +25947 -2093 |
| 候选×文件四态 | 524：ALREADY0 / CLEAN269 / CONFLICT199 / NOBASE56 |


| PR | merge | first parent | patch SHA256 |
|---|---|---|---|
| [#8174](https://github.com/Wei-Shaw/sub2api/pull/8174) | 2354c78288d05ec817b4f31a1cf5ebe231e36ba2 | 19d4cb222a24cecde72ba20045f7eb24cf2cf9a5 | 6f8ba84c67a4c6ed6d78e946c6813821e4138ef365341379bec315c71e1f53c4 |
| [#8577](https://github.com/Wei-Shaw/sub2api/pull/8577) | 19d4cb222a24cecde72ba20045f7eb24cf2cf9a5 | 9e9ca6fe95d3d20ca9d099b0c681e336e8d812c4 | dc40b70699b807fb2e125b35848073e421eb0a83e9dfbdbf6ac06cb5fecab064 |
| [#8653](https://github.com/Wei-Shaw/sub2api/pull/8653) | 83a1f5dbe82edf8b440d564811875c7460117837 | 314de5d0c09e73970a7cb930170f68f660245677 | dec12f6205ad149f3401a096e111ca2dfbe5ac6a1269ad5dfc2e36f6ad0b615d |
| [#8659](https://github.com/Wei-Shaw/sub2api/pull/8659) | 0c641e5ed7d6894d409915f5284652aa23eb8b72 | 83a1f5dbe82edf8b440d564811875c7460117837 | 20141f9529646f67c17f92c7748d140a4c5d369ec0a1205c7eeefe57c9c931bc |
| [#8662](https://github.com/Wei-Shaw/sub2api/pull/8662) | 980ddd942511262f49f10a51d304c1fa074d7cf7 | 0c641e5ed7d6894d409915f5284652aa23eb8b72 | 620f9b2d0b9fef1b5cee27621bc90d0b676e1be81b5dcc49fdd63cee6f1f1d10 |
| [#8667](https://github.com/Wei-Shaw/sub2api/pull/8667) | 41c567b38fb0a709f540f4c47477910522364eee | a6e02a37aafa07a2d149aa563120890a41ce255a | a3c371949d0b7dff00ddb21b3dedbcdb5d72ea3aa059de834475d6e8fdac9129 |
| [#8673](https://github.com/Wei-Shaw/sub2api/pull/8673) | 8da0388542f01bc5bf7317730d94fe3ce2c0aedf | 41c567b38fb0a709f540f4c47477910522364eee | d75310f1f741969981a318f6c509090b67e4abf240d88e8fd9713c19906eb939 |
| [#8675](https://github.com/Wei-Shaw/sub2api/pull/8675) | 0ebc8b2fd698739554271956df63119dfd79310f | 8da0388542f01bc5bf7317730d94fe3ce2c0aedf | b4d4e12d2c4595b35c220ce4c0736343da757f7cd34758e467280301e7ae6a49 |
| [#8724](https://github.com/Wei-Shaw/sub2api/pull/8724) | f8d891d6fbdfb3c378f827d897f45b80438d974f | 7a19bbef73b76e9158944d4a94c7e58bb5c78912 | a3940972eb2d86602a2baa83d805d62e10b83fb051a8a7fe35f024b86fd90408 |
| [#8727](https://github.com/Wei-Shaw/sub2api/pull/8727) | 32e4283894b45e2a7256acb374d0e4896bc6bd72 | 51e76a46c181ae93697f3a06702bc74ab4a45931 | b807a80da851be1195d80e6372521e9e77cf868cae3d3d52034980a425ed9be8 |
| [#8728](https://github.com/Wei-Shaw/sub2api/pull/8728) | a38ffc30506cdb48a1228082d2e2c86e199f5ddb | 32e4283894b45e2a7256acb374d0e4896bc6bd72 | 9c710050c9415ebfb5218f5ce5ddba9f34546a2eaa088f39e12164e53aff4af7 |
| [#8729](https://github.com/Wei-Shaw/sub2api/pull/8729) | c6d9d48f9a3b52d1ce7b48f020d4630162f8e486 | a38ffc30506cdb48a1228082d2e2c86e199f5ddb | 33e5fcb0661a61d41341c7208bd7b876ee4462054c52a5e0c96d52e2333ea5d9 |
| [#8737](https://github.com/Wei-Shaw/sub2api/pull/8737) | 5146728d2f98b73330f5c316807e495f44182a7c | 472ea5752598d7e4581ba195aaa72989c165640b | f482190b6de991b37acaccaa9a9c844e31f10a0bc52862de57f0e1296823ae52 |
| [#8739](https://github.com/Wei-Shaw/sub2api/pull/8739) | c782537eef5cee6d72614d963a815b4704039662 | d7ae0544e8f243b2956fbee6eaad823fbd35d03a | 55a13ef17edd991c8680e603dd6a6d822604287734ede6452de3a0ee2e195bcc |
| [#8741](https://github.com/Wei-Shaw/sub2api/pull/8741) | e54a186548a0ca1aac6211a7d8c0ed6f8b939d8f | 0563176a5d72c51a96a525322bd13b55faf6c6df | eb5705e1bf42d920d54e313075030500b525f2ed5a46d6b23d77693f480ac855 |
| [#8743](https://github.com/Wei-Shaw/sub2api/pull/8743) | 6247d33eb6009f557df3f7e1f23f3fbc74c9396e | 6fd905f30d890c18152c17050cba4232658a6c22 | ae4a7c8711651e876fdbe86a0a2bbfe680f41197fd440a771168dfed5d6c2416 |
| [#8745](https://github.com/Wei-Shaw/sub2api/pull/8745) | 8a585c91414150bc3ce31a18809d1760422c9614 | 6247d33eb6009f557df3f7e1f23f3fbc74c9396e | ac347315080bc2367cc60a46fd1bf0fff91df0ceb61d88846c28456944ae8cba |
| [#8747](https://github.com/Wei-Shaw/sub2api/pull/8747) | ef29203002bf134560b78627e675a79547b88ab3 | 7e4c614bcfb7fc93c955f6da3227c4c0cdfab8cb | 822ccb50551e4b0bfbdf5a2e3a21ece1b3809940e23ec8460abb8256aa5e1a38 |
| [#8754](https://github.com/Wei-Shaw/sub2api/pull/8754) | e16a349ebb102caef1ec3bf5804f0c7058251636 | 8a585c91414150bc3ce31a18809d1760422c9614 | 7de91f92e8e064f5ac117993f058ff05ed097df340c938db3a146cf807ac7524 |
| [#8767](https://github.com/Wei-Shaw/sub2api/pull/8767) | e3151c5201c2456a5cac09a3582de0fc1d7913a0 | 530c5e3a4f9c73ca2b0881191d0ef8e0d82c5484 | 9bed881da22e62fdf4d1455e0eb0546f29875d8fc74866b9b8f03752349e1ce5 |
| [#8772](https://github.com/Wei-Shaw/sub2api/pull/8772) | 3e8e3254d1a5c9f4e2bf66d05ca95ba57bdf6297 | f8d891d6fbdfb3c378f827d897f45b80438d974f | 2a7f1fd44638760de8b6962c761445f3cd4f8ab42b2a9f2a24fef33b9dccbade |
| [#8796](https://github.com/Wei-Shaw/sub2api/pull/8796) | 470c02b39944414f622615573bc00e0fe88275ca | 97957edd9e8e7db1388fb36d5e8d1143a59fa999 | 6723e263015ca132a5ca29f3c85a96c79a758c98c36f8b71eb231b52aa469aad |
| [#8798](https://github.com/Wei-Shaw/sub2api/pull/8798) | 857ab95081dad5a47132c5e4219fd6148391bedb | 470c02b39944414f622615573bc00e0fe88275ca | b0074e5d395e74a170d4eeb74d0dcaa4aa053ed429593d88734389f5c3efc753 |
| [#8800](https://github.com/Wei-Shaw/sub2api/pull/8800) | 119b04f945a055607f9f078d54dc31bf3dd8f1b6 | 857ab95081dad5a47132c5e4219fd6148391bedb | f6a2866406fb14f70b45592de2c2863bb16e9d988b60216e3798eabb2ead4ddd |
| [#8806](https://github.com/Wei-Shaw/sub2api/pull/8806) | 568db32f22c1e7b315367a2f5219565e0a9b776a | 119b04f945a055607f9f078d54dc31bf3dd8f1b6 | 317b6bacada7363dd4d0804fe0c2fe83a2987e5128d2c8e4f09edbdd19225fc9 |
| [#8811](https://github.com/Wei-Shaw/sub2api/pull/8811) | 1ca302abf36dd32963c9b15845be21b60720f38d | 9750f8e039a74bdb76cead0d1b23624db3f49ddb | 7d3f3c4f3c52c8bd6edc8eabaef93a1a00ceaa225d8f71f8c4e146cbf381589f |
| [#8813](https://github.com/Wei-Shaw/sub2api/pull/8813) | 101c2462ddb748e3ceb5e5a2e999fce9cf92467c | 1ca302abf36dd32963c9b15845be21b60720f38d | 6ead2b3f67a32a540aa609a450c821c72d07e8607be4a9a63c469fac985876cb |
| [#8822](https://github.com/Wei-Shaw/sub2api/pull/8822) | 04503e001e3d5ac3c156b8293accbebc31afeb17 | f242a6ea2327e2c849b8765bcfd3bf9e2abc7160 | 60afcbbd1b321186c1fe706bd6c1e91a3d088924abffa53aff47d98fd5b30561 |
| [#8826](https://github.com/Wei-Shaw/sub2api/pull/8826) | 4f3f9163172b27f29988811e82cea16024d5aaeb | f2eaed31c365da3b986eca4b28ed8c8ca53b530b | 8bcee8119f49185b4edc66c2907786181b6c1806359806de66724a43d8b2e20e |
| [#8828](https://github.com/Wei-Shaw/sub2api/pull/8828) | 46b879265896850f0ee3cb4971c5bdf3dde7df2f | 4f3f9163172b27f29988811e82cea16024d5aaeb | 59b835d32f2e885845fc93f278755cbf2b9eb229f8f8bd4f507b3f7e9774a58d |
| [#8834](https://github.com/Wei-Shaw/sub2api/pull/8834) | 4328510f285c2fbcd57fe4e57ab3a73a86e62368 | 03c17491b3ab69ff36ae7d45e15d3d6088be0c2d | 1509fb4de47e89dfecbddf68e6fc3291b31f32303f09b79456152d5e962c34d3 |
| [#8903](https://github.com/Wei-Shaw/sub2api/pull/8903) | 69d2fae91318359e4e25c0171eaf70364dad3629 | 45632495f593697393e6b52b3aabdab9eb1dd9ac | 74836b004fa76fac9f8d9ecd081f172cd4c7bc4dcaf260b319d77ba77422aa90 |
| [#8904](https://github.com/Wei-Shaw/sub2api/pull/8904) | 45632495f593697393e6b52b3aabdab9eb1dd9ac | a4d3103cd5a075505da604a3d2d5243cbc51f4ea | 5da9c5ec4ca4de24ccb84e091da977d88565531db61407ef04ba17cfa5c4c80b |
| [#8920](https://github.com/Wei-Shaw/sub2api/pull/8920) | 5af9afb588c4b014b3df9db7d4073af1f8ae01b6 | 2354c78288d05ec817b4f31a1cf5ebe231e36ba2 | 52e7d77855033c7d648c826d47a38a3cf1ef6b098f0f005ca7f4215c0d9e1e27 |


整PR first parent→merge为来源；最终tag状态优先于中间提交和发布摘要。全部文件先reverse后forward，未应用产品补丁。[完整来源与评估证据](../../../docs/upstream-sync/evidence-0.2.23/MANIFEST.md)。
