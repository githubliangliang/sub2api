import pathlib, shutil, subprocess, tempfile
root=pathlib.Path.cwd()
def run(*args, cwd=root):
    return subprocess.check_output(['rtk','proxy',*args],cwd=cwd,text=True).strip()
assert run('sh','backend/scripts/resolve-version.sh')=='1.1.20'
with tempfile.TemporaryDirectory(prefix='sub2api-v1.1.20-version-') as temp:
    fixture=pathlib.Path(temp)
    scripts=fixture/'backend/scripts'
    scripts.mkdir(parents=True)
    version=fixture/'backend/cmd/server/VERSION'
    version.parent.mkdir(parents=True)
    version.write_text('1.1.20\n')
    shutil.copy2(root/'backend/scripts/resolve-version.sh',scripts/'resolve-version.sh')
    run('git','init','-q',cwd=fixture)
    run('git','config','user.name','Release Test',cwd=fixture)
    run('git','config','user.email','release-test@example.invalid',cwd=fixture)
    run('git','add','.',cwd=fixture)
    run('git','commit','-qm','fixture',cwd=fixture)
    assert run('sh','backend/scripts/resolve-version.sh',cwd=fixture)=='1.1.20'
    run('git','tag','-a','v1.1.20','-m','v1.1.20 fixture',cwd=fixture)
    assert run('sh','backend/scripts/resolve-version.sh',cwd=fixture)=='1.1.20'
    version.write_text('1.1.21\n')
    assert run('sh','backend/scripts/resolve-version.sh',cwd=fixture)=='1.1.20'
    run('git','add','.',cwd=fixture)
    run('git','commit','-qm','next fixture',cwd=fixture)
    assert run('sh','backend/scripts/resolve-version.sh',cwd=fixture)=='1.1.21'
print('VERSION resolver: actual 1.1.20 fallback, isolated annotated tag priority and next commit fallback passed.')
